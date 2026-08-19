package server

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/podcctv/jijian/internal/model"
	"github.com/podcctv/jijian/internal/store"
)

type API struct {
	store   *store.Memory
	logger  *slog.Logger
	handler http.Handler
}

func New(st *store.Memory, logger *slog.Logger) *API {
	a := &API{store: st, logger: logger}
	mux := http.NewServeMux()
	a.routes(mux)
	a.handler = a.middleware(mux)
	return a
}
func (a *API) Handler() http.Handler { return a.handler }

func (a *API) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"status": "ok", "time": time.Now().UTC(), "version": "0.1.0"})
	})
	mux.HandleFunc("GET /api/v1/dashboard", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, a.store.Dashboard()) })
	mux.HandleFunc("GET /api/v1/nodes", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"items": a.store.Nodes()})
	})
	mux.HandleFunc("GET /api/v1/nodes/{id}", a.node)
	mux.HandleFunc("GET /api/v1/nodes/{id}/series", a.series)
	mux.HandleFunc("GET /api/v1/alerts", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"items": a.store.Alerts()})
	})
	mux.HandleFunc("POST /api/v1/nodes/{id}/scan", a.scan)
	mux.HandleFunc("POST /api/v1/enrollment-tokens", a.enrollment)
	mux.HandleFunc("POST /api/v1/agents/register", a.register)
	mux.HandleFunc("POST /api/v1/heartbeats", a.signed(a.heartbeat))
	mux.HandleFunc("POST /api/v1/reports", a.signed(a.report))
	mux.HandleFunc("GET /api/v1/agents/commands", a.signed(a.commands))
}

func (a *API) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Content-Digest, Signature-Input, Signature")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		start := time.Now()
		next.ServeHTTP(w, r)
		a.logger.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}

func (a *API) node(w http.ResponseWriter, r *http.Request) {
	n, err := a.store.Node(r.PathValue("id"))
	if err != nil {
		apiError(w, 404, "NODE_NOT_FOUND", "node not found")
		return
	}
	writeJSON(w, 200, n)
}
func (a *API) series(w http.ResponseWriter, r *http.Request) {
	s, err := a.store.Series(r.PathValue("id"))
	if err != nil {
		apiError(w, 404, "NODE_NOT_FOUND", "node not found")
		return
	}
	writeJSON(w, 200, map[string]any{"node_id": r.PathValue("id"), "metric": "risk_score", "step": "12h", "series": s})
}
func (a *API) scan(w http.ResponseWriter, r *http.Request) {
	c, err := a.store.CreateScan(r.PathValue("id"))
	if err != nil {
		apiError(w, 404, "NODE_NOT_FOUND", "node not found")
		return
	}
	writeJSON(w, 202, map[string]any{"accepted": true, "command": c})
}

func (a *API) enrollment(w http.ResponseWriter, r *http.Request) {
	var in struct{ Name, Provider, Region string }
	if err := decode(r, &in, 32<<10); err != nil || strings.TrimSpace(in.Name) == "" {
		apiError(w, 400, "INVALID_PAYLOAD", "name is required")
		return
	}
	e := a.store.CreateEnrollment("tenant_demo", in.Name, in.Provider, in.Region)
	writeJSON(w, 201, map[string]any{"token": e.Token, "expires_at": e.ExpiresAt, "max_uses": 1})
}
func (a *API) register(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	var in struct {
		PublicKey string `json:"public_key"`
	}
	if err := decode(r, &in, 32<<10); err != nil {
		apiError(w, 400, "INVALID_PAYLOAD", err.Error())
		return
	}
	pub, err := base64.StdEncoding.DecodeString(in.PublicKey)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		apiError(w, 400, "INVALID_PUBLIC_KEY", "public_key must be base64 Ed25519 key")
		return
	}
	n, agent, err := a.store.Register(token, pub)
	if err != nil {
		apiError(w, 401, "ENROLLMENT_EXPIRED", "token is invalid, used, or expired")
		return
	}
	writeJSON(w, 201, map[string]any{"agent_id": agent.AgentID, "node_id": n.ID, "tenant_id": n.TenantID})
}

type signedHandler func(http.ResponseWriter, *http.Request, []byte, store.AgentKey)

func (a *API) signed(next signedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 2<<20))
		if err != nil {
			apiError(w, 400, "INVALID_PAYLOAD", "unable to read body")
			return
		}
		meta, err := parseSignatureInput(r.Header.Get("Signature-Input"))
		if err != nil {
			apiError(w, 401, "INVALID_SIGNATURE", err.Error())
			return
		}
		agent, err := a.store.Agent(meta.KeyID)
		if err != nil {
			apiError(w, 401, "INVALID_SIGNATURE", "unknown keyid")
			return
		}
		if err = verifySignature(r, body, agent.PublicKey, meta); err != nil {
			apiError(w, 401, "INVALID_SIGNATURE", err.Error())
			return
		}
		if !a.store.UseNonce(meta.KeyID, meta.Nonce, 10*time.Minute) {
			apiError(w, 409, "REPLAY_DETECTED", "nonce has already been used")
			return
		}
		next(w, r, body, agent)
	}
}
func (a *API) heartbeat(w http.ResponseWriter, r *http.Request, body []byte, agent store.AgentKey) {
	var h model.Heartbeat
	if err := json.Unmarshal(body, &h); err != nil {
		apiError(w, 400, "INVALID_PAYLOAD", err.Error())
		return
	}
	if h.AgentID != agent.AgentID || h.NodeID != agent.NodeID {
		apiError(w, 403, "AGENT_FORBIDDEN", "agent identity does not match payload")
		return
	}
	if h.ObservedAt.IsZero() {
		h.ObservedAt = time.Now().UTC()
	}
	if err := a.store.SaveHeartbeat(h); err != nil {
		apiError(w, 404, "NODE_NOT_FOUND", err.Error())
		return
	}
	writeJSON(w, 202, map[string]any{"accepted": true})
}
func (a *API) report(w http.ResponseWriter, r *http.Request, body []byte, agent store.AgentKey) {
	var report model.Report
	if err := json.Unmarshal(body, &report); err != nil {
		apiError(w, 400, "INVALID_PAYLOAD", err.Error())
		return
	}
	if report.SchemaVersion != "1.0" {
		apiError(w, 422, "SCHEMA_UNSUPPORTED", "only schema_version 1.0 is supported")
		return
	}
	if report.AgentID != agent.AgentID || report.NodeID != agent.NodeID {
		apiError(w, 403, "AGENT_FORBIDDEN", "agent identity does not match payload")
		return
	}
	if report.ReportID == "" || report.CollectedAt.IsZero() || report.Network.Family != 4 && report.Network.Family != 6 {
		apiError(w, 400, "INVALID_PAYLOAD", "report_id, collected_at and valid network.family are required")
		return
	}
	if err := a.store.SaveReport(report); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			apiError(w, 409, "DUPLICATE_REPORT", err.Error())
			return
		}
		apiError(w, 400, "INVALID_PAYLOAD", err.Error())
		return
	}
	writeJSON(w, 202, map[string]any{"accepted": true, "report_id": report.ReportID})
}
func (a *API) commands(w http.ResponseWriter, r *http.Request, body []byte, agent store.AgentKey) {
	writeJSON(w, 200, map[string]any{"items": a.store.Commands(agent.AgentID)})
}

func decode(r *http.Request, dst any, max int64) error {
	d := json.NewDecoder(io.LimitReader(r.Body, max))
	d.DisallowUnknownFields()
	return d.Decode(dst)
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func apiError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message, "request_id": fmt.Sprintf("req_%d", time.Now().UnixNano())}})
}

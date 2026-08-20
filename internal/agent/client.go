package agent

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/podcctv/detective-chicken/internal/model"
)

const Version = "0.1.0"

type Config struct {
	ServerURL  string `json:"server_url"`
	AgentID    string `json:"agent_id"`
	NodeID     string `json:"node_id"`
	TenantID   string `json:"tenant_id"`
	PrivateKey string `json:"private_key"`
}

type Client struct {
	Config Config
	HTTP   *http.Client
}

type Directive struct {
	ScanDue             bool `json:"scan_due"`
	ScanIntervalMinutes int  `json:"scan_interval_minutes"`
	Commands            []struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"commands"`
}

func Enroll(serverURL, token, configPath string) (Config, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return Config{}, err
	}
	body, _ := json.Marshal(map[string]string{"public_key": base64.StdEncoding.EncodeToString(pub)})
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(serverURL, "/")+"/api/v1/agents/register", bytes.NewReader(body))
	if err != nil {
		return Config{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Config{}, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode != 201 {
		return Config{}, fmt.Errorf("registration failed: %s: %s", res.Status, raw)
	}
	var out struct {
		AgentID  string `json:"agent_id"`
		NodeID   string `json:"node_id"`
		TenantID string `json:"tenant_id"`
	}
	if err = json.Unmarshal(raw, &out); err != nil {
		return Config{}, err
	}
	cfg := Config{ServerURL: strings.TrimRight(serverURL, "/"), AgentID: out.AgentID, NodeID: out.NodeID, TenantID: out.TenantID, PrivateKey: base64.StdEncoding.EncodeToString(priv)}
	if err = SaveConfig(configPath, cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func LoadConfig(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err = json.Unmarshal(raw, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.ServerURL == "" || cfg.AgentID == "" || cfg.NodeID == "" || cfg.PrivateKey == "" {
		return Config{}, errors.New("incomplete agent config")
	}
	return cfg, nil
}
func SaveConfig(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	raw, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(path, raw, 0600)
}

func (c *Client) privateKey() (ed25519.PrivateKey, error) {
	raw, err := base64.StdEncoding.DecodeString(c.Config.PrivateKey)
	if err != nil || len(raw) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid private key in config")
	}
	return ed25519.PrivateKey(raw), nil
}

func (c *Client) signed(method, path string, payload any) (*http.Response, error) {
	body := []byte{}
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}
	priv, err := c.privateKey()
	if err != nil {
		return nil, err
	}
	created := time.Now().Unix()
	nonceBytes := make([]byte, 16)
	_, _ = rand.Read(nonceBytes)
	nonce := base64.RawURLEncoding.EncodeToString(nonceBytes)
	sum := sha256.Sum256(body)
	digest := "sha-256=:" + base64.StdEncoding.EncodeToString(sum[:]) + ":"
	base := fmt.Sprintf("%s\n%s\n%s\n%d\n%s\n%s", strings.ToUpper(method), path, digest, created, nonce, c.Config.AgentID)
	sig := ed25519.Sign(priv, []byte(base))
	req, err := http.NewRequest(method, c.Config.ServerURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Digest", digest)
	req.Header.Set("Signature-Input", `sig1=("@method" "@path" "content-digest");created=`+strconv.FormatInt(created, 10)+`;keyid="`+c.Config.AgentID+`";nonce="`+nonce+`"`)
	req.Header.Set("Signature", "sig1=:"+base64.StdEncoding.EncodeToString(sig)+":")
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}
	return httpClient.Do(req)
}

func (c *Client) Heartbeat(status map[string]any) (Directive, error) {
	if status == nil {
		status = map[string]any{"state": "ready"}
	}
	h := model.Heartbeat{AgentID: c.Config.AgentID, NodeID: c.Config.NodeID, ObservedAt: time.Now().UTC(), AgentVersion: Version, Status: status}
	res, err := c.signed(http.MethodPost, "/api/v1/heartbeats", h)
	if err != nil {
		return Directive{}, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return Directive{}, fmt.Errorf("server returned %s: %s", res.Status, body)
	}
	var directive Directive
	if err := json.Unmarshal(body, &directive); err != nil {
		return Directive{}, err
	}
	return directive, nil
}
func (c *Client) Upload(report model.Report) error {
	report.AgentID = c.Config.AgentID
	report.NodeID = c.Config.NodeID
	return c.post("/api/v1/reports", report)
}
func (c *Client) post(path string, payload any) error {
	res, err := c.signed(http.MethodPost, path, payload)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
		return fmt.Errorf("server returned %s: %s", res.Status, body)
	}
	return nil
}

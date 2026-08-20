package server

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed install_template.sh
var installTemplate string

func publicBaseURL(r *http.Request) string {
	if configured := strings.TrimRight(os.Getenv("DETECTIVE_CHICKEN_PUBLIC_URL"), "/"); configured != "" {
		return configured
	}
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	host := r.Host
	if forwarded := r.Header.Get("X-Forwarded-Host"); forwarded != "" {
		host = forwarded
	}
	return scheme + "://" + host
}

func choice(value, fallback string, allowed ...string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, candidate := range allowed {
		if value == candidate {
			return value
		}
	}
	return fallback
}

func (a *API) installScript(w http.ResponseWriter, r *http.Request) {
	file := r.PathValue("file")
	if !strings.HasSuffix(file, ".sh") {
		apiError(w, http.StatusNotFound, "INSTALLER_NOT_FOUND", "installer not found")
		return
	}
	token := strings.TrimSuffix(file, ".sh")
	enrollment, err := a.store.Enrollment(token)
	if err != nil {
		apiError(w, http.StatusNotFound, "INSTALLER_EXPIRED", "installer token is invalid, used, or expired")
		return
	}
	data := struct{ ServerURL, Token, OSFamily, Platform, Arch string }{
		ServerURL: publicBaseURL(r), Token: enrollment.Token,
		OSFamily: choice(enrollment.OSFamily, "auto", "auto", "debian", "alpine", "rhel", "arch"),
		Platform: choice(enrollment.Platform, "auto", "auto", "pve", "baremetal", "lxc", "docker", "podman", "incus"),
		Arch:     choice(enrollment.Arch, "auto", "auto", "amd64", "arm64", "armv7"),
	}
	tmpl, err := template.New("installer").Parse(installTemplate)
	if err != nil {
		apiError(w, http.StatusInternalServerError, "INSTALLER_FAILED", "installer template is invalid")
		return
	}
	w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	w.Header().Set("Content-Disposition", `inline; filename="install-detective-chicken.sh"`)
	if err := tmpl.Execute(w, data); err != nil {
		a.logger.Error("render installer", "error", err)
	}
}

func (a *API) agentDownload(w http.ResponseWriter, r *http.Request) {
	arch := choice(r.PathValue("arch"), "", "amd64", "arm64", "armv7")
	if arch == "" {
		apiError(w, http.StatusNotFound, "ARCH_UNSUPPORTED", "supported architectures: amd64, arm64, armv7")
		return
	}
	dir := os.Getenv("DETECTIVE_CHICKEN_AGENT_DIR")
	if dir == "" {
		dir = "/usr/local/share/detective-chicken"
	}
	path := filepath.Join(dir, "detective-chicken-agent-"+arch)
	if _, err := os.Stat(path); err != nil {
		apiError(w, http.StatusNotFound, "AGENT_NOT_BUILT", fmt.Sprintf("agent binary for %s is unavailable", arch))
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="detective-chicken-agent"`)
	http.ServeFile(w, r, path)
}

// Package gateway is the inference data plane: the HTTP surface OpenCode sees as the
// model provider named "agentpool" (spec v5.2 §3, §9.2).
//
// Status: SKELETON. It enforces the listener and request-origin rules that every later
// milestone depends on, and answers every inference path with an explicit
// not-implemented error. It never fakes a model response (P1 gate: "no hidden fake success").
package gateway

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Nicolas1bhr/agentpool/internal/version"
)

// DefaultAddr is the reference loopback bind from the spec's route manifest example (§13.3).
const DefaultAddr = "127.0.0.1:43819"

// CheckLoopback rejects any bind address that is not a literal loopback IP.
//
// v1 has no LAN/WAN listener (§9.2). A hostname such as "localhost" is refused because its
// resolution is not under AgentPool's control, and an empty host would bind every interface.
func CheckLoopback(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid listen address %q: %w", addr, err)
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("refusing to listen on %q: v1 binds a literal loopback IP only (127.0.0.1 or ::1)", addr)
	}
	return nil
}

// NewServer returns an HTTP server for the data plane with bounded timeouts.
// It fails closed if addr is not loopback.
func NewServer(addr string) (*http.Server, error) {
	if err := CheckLoopback(addr); err != nil {
		return nil, err
	}
	return &http.Server{
		Addr:              addr,
		Handler:           NewHandler(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    64 << 10,
	}, nil
}

// NewHandler returns the data-plane HTTP handler.
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"version": version.Version,
			"stage":   "skeleton: inference not implemented",
		})
	})
	mux.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "agentpool_not_implemented",
			fmt.Sprintf("%s %s is not implemented yet; this AgentPool build is a skeleton (see ROADMAP.md, milestone P1)", r.Method, r.URL.Path))
	})
	return rejectBrowserOrigin(mux)
}

// rejectBrowserOrigin refuses any request carrying an Origin header. Loopback is reachable from
// web pages in the owner's browser; browser-origin requests are rejected and CORS is never
// enabled (§9.2).
func rejectBrowserOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") != "" {
			writeError(w, http.StatusForbidden, "agentpool_browser_origin_rejected",
				"browser-origin requests are not accepted by the AgentPool data plane")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// errorBody mirrors the OpenAI-compatible error envelope OpenCode's provider packages parse.
type errorBody struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	var body errorBody
	body.Error.Message = msg
	body.Error.Type = strings.ToLower(strings.ReplaceAll(http.StatusText(status), " ", "_"))
	body.Error.Code = code
	writeJSON(w, status, body)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

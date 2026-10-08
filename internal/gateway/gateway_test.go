package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCheckLoopback(t *testing.T) {
	cases := []struct {
		addr string
		ok   bool
	}{
		{"127.0.0.1:43819", true},
		{"[::1]:43819", true},
		{"127.0.0.2:43819", true}, // all of 127.0.0.0/8 is loopback
		{":43819", false},         // every interface
		{"0.0.0.0:43819", false},
		{"[::]:43819", false},
		{"192.168.1.10:43819", false},
		{"localhost:43819", false}, // resolution not under our control
		{"127.0.0.1", false},       // no port
	}
	for _, c := range cases {
		err := CheckLoopback(c.addr)
		if (err == nil) != c.ok {
			t.Errorf("CheckLoopback(%q) error = %v, want ok=%v", c.addr, err, c.ok)
		}
	}
}

func TestNewServerFailsClosed(t *testing.T) {
	if _, err := NewServer("0.0.0.0:43819"); err == nil {
		t.Fatal("NewServer accepted a non-loopback address")
	}
	srv, err := NewServer(DefaultAddr)
	if err != nil {
		t.Fatalf("NewServer(DefaultAddr): %v", err)
	}
	if srv.ReadHeaderTimeout == 0 {
		t.Error("server has no ReadHeaderTimeout")
	}
}

func TestHealthz(t *testing.T) {
	rec := serve(t, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Errorf("status field = %q", body["status"])
	}
}

// The skeleton must never fake a model response: every inference path is an explicit 501.
func TestInferencePathsAreHonestlyUnimplemented(t *testing.T) {
	for _, path := range []string{"/v1/chat/completions", "/v1/responses", "/v1/models"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"model":"x","stream":true}`))
		rec := serve(t, req)
		if rec.Code != http.StatusNotImplemented {
			t.Errorf("%s: status = %d, want 501", path, rec.Code)
		}
		var body errorBody
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("%s: error body is not JSON: %v", path, err)
		}
		if body.Error.Code != "agentpool_not_implemented" {
			t.Errorf("%s: error code = %q", path, body.Error.Code)
		}
	}
}

func TestBrowserOriginRejected(t *testing.T) {
	for _, path := range []string{"/healthz", "/v1/chat/completions"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Origin", "https://evil.example")
		rec := serve(t, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s with Origin: status = %d, want 403", path, rec.Code)
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("%s: CORS header set to %q", path, got)
		}
	}
}

func TestUnknownPathNotFound(t *testing.T) {
	rec := serve(t, httptest.NewRequest(http.MethodGet, "/admin", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func serve(t *testing.T, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	NewHandler().ServeHTTP(rec, req)
	return rec
}

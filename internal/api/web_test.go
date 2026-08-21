package api

import (
	"net/http"
	"strings"
	"testing"
)

func TestLandingPageServesHTML(t *testing.T) {
	st, _ := newAPITestStore(t)
	router := New(st).Router()

	rec := performJSON(t, router, http.MethodGet, "/", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want text/html; charset=utf-8", got)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "MeshDNS") {
		t.Fatalf("body missing MeshDNS: %s", body)
	}
	if !strings.Contains(body, "/v0/resolve") {
		t.Fatalf("body missing /v0/resolve: %s", body)
	}
}

func TestLandingPageAndStatsCoexist(t *testing.T) {
	st, _ := newAPITestStore(t)
	router := New(st).Router()

	rec := performJSON(t, router, http.MethodGet, "/v0/stats", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	if !strings.Contains(rec.Body.String(), "\"servers_active\"") {
		t.Fatalf("stats body missing JSON payload: %s", rec.Body.String())
	}
}

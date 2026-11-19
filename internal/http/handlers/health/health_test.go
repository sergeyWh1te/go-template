package health_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sergeyWh1te/go-template/internal/http/handlers/health"
)

func TestHandler_returns_status_ok(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", http.NoBody)
	rec := httptest.NewRecorder()

	health.New().Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if body := rec.Body.String(); body != `{"status":"ok"}` {
		t.Errorf(`expected {"status":"ok"}, got %s`, body)
	}
}

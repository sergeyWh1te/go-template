package userexample_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	userexample "github.com/sergeyWh1te/go-template/internal/http/handlers/user_example"
	"github.com/sergeyWh1te/go-template/internal/pkg/users"
	"github.com/sergeyWh1te/go-template/internal/pkg/users/entity"
	"github.com/sergeyWh1te/go-template/internal/pkg/users/mocks"
)

// serve runs one GET /example through the handler with a mocked usecase and
// returns the recorded response plus everything that was logged, as JSON lines.
func serve(t *testing.T, user *entity.User, err error) (rec *httptest.ResponseRecorder, logs string) {
	t.Helper()

	uc := mocks.NewUsecase(t)
	uc.EXPECT().Get(mock.Anything, int64(1)).Return(user, err).Once()

	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/example", http.NoBody)
	rec = httptest.NewRecorder()

	userexample.New(log, uc).Handler(rec, req)

	return rec, buf.String()
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body %q: %v", rec.Body.String(), err)
	}

	return body.Error
}

func TestHandler_returns_200_with_the_user(t *testing.T) {
	rec, logs := serve(t, &entity.User{ID: 1}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var user entity.User
	if err := json.Unmarshal(rec.Body.Bytes(), &user); err != nil {
		t.Fatalf("decode body %q: %v", rec.Body.String(), err)
	}

	if user.ID != 1 {
		t.Errorf("expected user id 1, got %d", user.ID)
	}

	if logs != "" {
		t.Errorf("expected nothing logged on success, got %s", logs)
	}
}

func TestHandler_returns_404_when_the_user_is_not_found(t *testing.T) {
	rec, logs := serve(t, nil, fmt.Errorf("get user 1: %w", users.ErrNotFound))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	if msg := decodeError(t, rec); msg != "user not found" {
		t.Errorf(`expected error "user not found", got %q`, msg)
	}

	if !strings.Contains(logs, `"level":"WARN"`) {
		t.Errorf("expected a WARN log line, got %s", logs)
	}
}

func TestHandler_returns_500_on_any_other_error(t *testing.T) {
	rec, logs := serve(t, nil, errors.New("connection refused"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	if msg := decodeError(t, rec); msg != "internal server error" {
		t.Errorf(`expected error "internal server error", got %q`, msg)
	}

	if !strings.Contains(logs, `"level":"ERROR"`) {
		t.Errorf("expected an ERROR log line, got %s", logs)
	}

	// The cause is logged for operators but never leaked to the client.
	if strings.Contains(rec.Body.String(), "connection refused") {
		t.Errorf("internal error leaked into the response: %s", rec.Body.String())
	}
}

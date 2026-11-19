package userexample

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sergeyWh1te/go-template/internal/pkg/users"
	"github.com/sergeyWh1te/go-template/internal/utils/deps"
)

type handler struct {
	log    deps.Logger
	userUc users.Usecase
}

type errorResponse struct {
	Error string `json:"error"`
}

func New(log deps.Logger, userUc users.Usecase) *handler {
	return &handler{
		log:    log,
		userUc: userUc,
	}
}

func (h *handler) Handler(w http.ResponseWriter, r *http.Request) {
	user, err := h.userUc.Get(r.Context(), int64(1))
	if errors.Is(err, users.ErrNotFound) {
		h.log.Warn("get user", "err", err)
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if err != nil {
		h.log.Error("get user", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	jsonResponse, err := json.Marshal(user)
	if err != nil {
		h.log.Error("marshal user", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonResponse)
}

// writeError writes {"error": msg}. Marshaling a single string field cannot
// fail, so the error is deliberately dropped.
func writeError(w http.ResponseWriter, status int, msg string) {
	jsonResponse, _ := json.Marshal(errorResponse{Error: msg})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(jsonResponse)
}

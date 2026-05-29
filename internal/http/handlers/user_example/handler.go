package userexample

import (
	"encoding/json"
	"net/http"

	"github.com/sergeyWh1te/go-template/internal/pkg/users"
	"github.com/sergeyWh1te/go-template/internal/utils/deps"
)

type handler struct {
	log    deps.Logger
	userUc users.Usecase
}

func New(log deps.Logger, userUc users.Usecase) *handler {
	return &handler{
		log:    log,
		userUc: userUc,
	}
}

func (h *handler) Handler(w http.ResponseWriter, r *http.Request) {
	user, err := h.userUc.Get(r.Context(), int64(1))
	if err != nil {
		h.log.Error("get user", "err", err)
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	jsonResponse, err := json.Marshal(user)
	if err != nil {
		h.log.Error("marshal user", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonResponse)
}

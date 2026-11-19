package usecase_test

import (
	"errors"
	"testing"

	"github.com/sergeyWh1te/go-template/internal/pkg/users"
	"github.com/sergeyWh1te/go-template/internal/pkg/users/entity"
	"github.com/sergeyWh1te/go-template/internal/pkg/users/mocks"
	"github.com/sergeyWh1te/go-template/internal/pkg/users/usecase"
)

func TestUsecase_Get_returns_the_repository_user(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().Get(t.Context(), int64(7)).Return(&entity.User{ID: 7}, nil).Once()

	user, err := usecase.New(repo).Get(t.Context(), 7)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}

	if user == nil || user.ID != 7 {
		t.Fatalf("expected user 7, got %+v", user)
	}
}

func TestUsecase_Get_passes_ErrNotFound_through(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().Get(t.Context(), int64(404)).Return(nil, users.ErrNotFound).Once()

	user, err := usecase.New(repo).Get(t.Context(), 404)
	if !errors.Is(err, users.ErrNotFound) {
		t.Fatalf("expected users.ErrNotFound, got %v", err)
	}

	if user != nil {
		t.Errorf("expected a nil user alongside the error, got %+v", user)
	}
}

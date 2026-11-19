package users

import (
	"context"

	"github.com/sergeyWh1te/go-template/internal/pkg/users/entity"
)

type Repository interface {
	Get(ctx context.Context, id int64) (*entity.User, error)
	Create(ctx context.Context) (*int64, error)
}

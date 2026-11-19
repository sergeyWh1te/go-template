package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/sergeyWh1te/go-template/internal/pkg/users"
	"github.com/sergeyWh1te/go-template/internal/pkg/users/entity"
)

type repo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) users.Repository {
	return &repo{
		db: db,
	}
}

func (r *repo) Get(ctx context.Context, id int64) (*entity.User, error) {
	var out entity.User

	err := r.db.GetContext(ctx, &out, `select * from users where id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get user %d: %w", id, users.ErrNotFound)
	}

	if err != nil {
		return nil, fmt.Errorf("get user %d: %w", id, err)
	}

	return &out, nil
}

func (r *repo) Create(ctx context.Context) (*int64, error) {
	var id int64

	query := `insert into users default values returning id;`
	if createUserErr := r.db.GetContext(ctx, &id, query); createUserErr != nil {
		return nil, fmt.Errorf("create user: %w", createUserErr)
	}

	return &id, nil
}

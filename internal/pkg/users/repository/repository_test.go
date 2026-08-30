package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/sergeyWh1te/go-template/internal/pkg/users/repository"
	"github.com/sergeyWh1te/go-template/internal/utils/testdb"
)

func TestRepository_Create_returns_a_new_id(t *testing.T) {
	db := testdb.New(t)
	repo := repository.New(db)

	id, err := repo.Create(context.Background())
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if id == nil {
		t.Fatal("expected an id, got nil")
	}

	// Truncate restarts identity, so the first row of a clean schema is id 1.
	if *id != 1 {
		t.Errorf("expected id 1 on a truncated table, got %d", *id)
	}
}

func TestRepository_Get_returns_the_created_user(t *testing.T) {
	db := testdb.New(t)
	repo := repository.New(db)

	ctx := context.Background()

	id, err := repo.Create(ctx)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	user, err := repo.Get(ctx, *id)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}

	if user.ID != *id {
		t.Errorf("expected user id %d, got %d", *id, user.ID)
	}
}

func TestRepository_Get_missing_user_returns_ErrNoRows(t *testing.T) {
	db := testdb.New(t)
	repo := repository.New(db)

	_, err := repo.Get(context.Background(), 404)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for a missing user, got %v", err)
	}
}

// Each test gets a truncated schema; this asserts that isolation rather than
// assuming it, since every other test depends on it.
func TestRepository_each_test_starts_from_a_clean_table(t *testing.T) {
	db := testdb.New(t)
	repo := repository.New(db)

	id, err := repo.Create(context.Background())
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if *id != 1 {
		t.Errorf("table was not clean: expected id 1, got %d", *id)
	}
}

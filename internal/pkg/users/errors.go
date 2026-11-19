package users

import "errors"

// ErrNotFound is returned (wrapped) by Repository.Get when no user has the
// requested id. Repositories translate driver errors such as sql.ErrNoRows
// into it, so callers check errors.Is(err, users.ErrNotFound) and never import
// the database driver.
var ErrNotFound = errors.New("user not found")

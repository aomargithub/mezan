package db

import (
	"errors"
)

var (
	ErrNoRecord       = errors.New("no matching record found")
	ErrDuplicateEmail = errors.New("duplicate user email")
)

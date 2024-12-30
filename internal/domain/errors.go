package domain

import (
	"errors"
)

var (
	ErrNoRecord        = errors.New("no matching record found")
	ErrDuplicateRecord = errors.New("duplicate record")
	ErrDuplicateEmail  = errors.New("duplicate user email")
)

package db

import (
	"errors"
)


var  ErrNoRecord = errors.New("no matching record found")
package domain

import (
	"time"

	"github.com/google/uuid"
)

type Mezani struct {
	Id uuid.UUID
	Name string
	CreatedAt time.Time
}
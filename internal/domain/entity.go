package domain

import (
	"time"
)

type Mezani struct {
	Id             int
	Name           string
	CreatorId      int
	TotalAmount    float32
	SettledPercent float32
	LastUpdatedAt  time.Time
	CreatedAt      time.Time
}

type User struct {
	Id        int
	Name      string
	Email     string
	CreatedAt time.Time
}

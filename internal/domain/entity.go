package domain

import (
	"time"
)

type Mezani struct {
	Id            int
	Name          string
	Creator       User
	TotalAmount   float32
	SettledAmount float32
	LastUpdatedAt time.Time
	CreatedAt     time.Time
	ShareId       string
	Expenses      []Expense
}

type User struct {
	Id        int
	Name      string
	Email     string
	CreatedAt time.Time
}

type MemberShip struct {
	Id        int
	Mezani    Mezani
	Member    User
	CreatedAt time.Time
}

type Expense struct {
	Id            int
	Name          string
	Creator       User
	Mezani        Mezani
	TotalAmount   float32
	SettledAmount float32
	LastUpdatedAt time.Time
	CreatedAt     time.Time
	Items         []ExpenseItem
}

type ExpenseItem struct {
	Id            int
	Name          string
	Creator       User
	Mezani        Mezani
	Amount        float32
	TotalAmount   float32
	SettledAmount float32
	LastUpdatedAt time.Time
	CreatedAt     time.Time
	Expense       Expense
	Quantity      float32
}

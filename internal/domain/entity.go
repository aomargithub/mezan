package domain

import (
	"slices"
	"time"
)

type Mezani struct {
	Id              int
	Name            string
	Creator         User
	TotalAmount     float32
	AllocatedAmount float32
	LastUpdatedAt   time.Time
	CreatedAt       time.Time
	ShareId         string
	Expenses        []Expense
	HasExpenses     bool
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
	Id              int
	Name            string
	Creator         User
	Mezani          Mezani
	TotalAmount     float32
	AllocatedAmount float32
	LastUpdatedAt   time.Time
	CreatedAt       time.Time
	Items           []ExpenseItem
	HasItems        bool
}

type ExpenseItem struct {
	Id              int
	Name            string
	Creator         User
	Mezani          Mezani
	Amount          float32
	TotalAmount     float32
	AllocatedAmount float32
	LastUpdatedAt   time.Time
	CreatedAt       time.Time
	Expense         Expense
	Quantity        float32
}
type ShareType string

func NewShareType(value string) *ShareType {
	st := ShareType(value)
	if slices.Contains(ShareTypes, st) {
		return &st
	}
	return nil
}

const (
	PERCENTAGE = ShareType("PERCENTAGE")
	EXACT      = ShareType("EXACT")
)

var ShareTypes = []ShareType{PERCENTAGE, EXACT}

type MezaniShare struct {
	Id          int
	CreatedAt   time.Time
	Share       float32
	ShareType   ShareType
	Amount      float32
	Mezani      Mezani
	Participant User
}

type ExpenseShare struct {
	Id          int
	CreatedAt   time.Time
	Share       float32
	ShareType   ShareType
	Amount      float32
	Mezani      Mezani
	Expense     Expense
	Participant User
}

type ExpenseItemShare struct {
	Id          int
	CreatedAt   time.Time
	Share       float32
	ShareType   ShareType
	Amount      float32
	Mezani      Mezani
	Expense     Expense
	ExpenseItem ExpenseItem
	Participant User
}

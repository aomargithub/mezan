package db

import "database/sql"

type ExpenseItemShareService struct {
	DB *sql.DB
	dbCommons
}

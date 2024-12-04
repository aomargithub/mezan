package db

import (
	"database/sql"
	"github.com/aomargithub/mezan/internal/domain"
)

type ExpenseItemService struct {
	DB *sql.DB
}

func (e ExpenseItemService) Create(expenseItem domain.ExpenseItem) error {
	stmt := `insert into expense_items (name, amount, mezani_id, expense_id, creator_id, created_at) values($1,$2,$3,$4,$5,$6)`
	_, err := e.DB.Exec(stmt, expenseItem.Name, expenseItem.Amount,
		expenseItem.Mezani.Id, expenseItem.Expense.Id, expenseItem.Creator.Id,
		expenseItem.CreatedAt)
	return err
}

package db

import (
	"database/sql"
	"github.com/aomargithub/mezan/internal/domain"
)

type ExpenseItemService struct {
	DB *sql.DB
	dbCommons
}

func (e ExpenseItemService) Create(expenseItem domain.ExpenseItem) error {
	stmt := `insert into expense_items (name, amount, total_amount, quantity, mezani_id, expense_id, creator_id, created_at) values($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := e.DB.Exec(stmt, expenseItem.Name, expenseItem.Amount, expenseItem.TotalAmount,
		expenseItem.Quantity, expenseItem.Mezani.Id, expenseItem.Expense.Id, expenseItem.Creator.Id,
		expenseItem.CreatedAt)
	return err
}

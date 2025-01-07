package db

import (
	"database/sql"
	"errors"
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

func (e ExpenseItemService) Get(expenseItemId int) (domain.ExpenseItem, error) {
	var item domain.ExpenseItem
	stmt := `select ei.id,
				   ei.name,
				   ei.created_at,
				   ei.allocated_amount,
				   ei.amount,
				   ei.quantity,
				   ei.total_amount,
				   u1.name
			from expense_items ei
					 join users u1 on u1.id = ei.creator_id
			where ei.id = $1;`
	row := e.DB.QueryRow(stmt, expenseItemId)
	err := row.Scan(&item.Id, &item.Name, &item.CreatedAt, &item.AllocatedAmount, &item.Amount, &item.Quantity,
		&item.TotalAmount, &item.Creator.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ExpenseItem{}, domain.ErrNoRecord
		}
		return domain.ExpenseItem{}, err
	}
	return item, err
}

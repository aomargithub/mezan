package db

import (
	"database/sql"
	"errors"
	"time"

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

func (e ExpenseItemService) Update(expenseItem domain.ExpenseItem) error {
	stmt := `update expense_items set name = $1, amount = $2, total_amount = $3, quantity = $4, last_updated_at = $5 where id = $6`
	_, err := e.DB.Exec(stmt, expenseItem.Name, expenseItem.Amount, expenseItem.TotalAmount,
		expenseItem.Quantity, expenseItem.LastUpdatedAt, expenseItem.Id)
	return err
}

func (e ExpenseItemService) GetExpenseIdMezaniIdTotalAmount(expenseItemId int) (int, int, float32, error) {
	stmt := " select mezani_id, expense_id, total_amount from expense_items where id = $1"

	r := e.DB.QueryRow(stmt, expenseItemId)
	var mezaniId, expenseId int
	var totalAmount float32
	err := r.Scan(&mezaniId, &expenseId, &totalAmount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, 0, domain.ErrNoRecord
		}
		return 0, 0, 0, err
	}
	return mezaniId, expenseId, totalAmount, nil
}

func (e ExpenseItemService) GetTotalAndAllocatedAmounts(expenseItemId int) (float32, float32, error) {
	var totalAmount, allocatedAmount float32
	stmt := `select total_amount, allocated_amount from expense_items where id = $1`
	row := e.DB.QueryRow(stmt, expenseItemId)
	err := row.Scan(&totalAmount, &allocatedAmount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, domain.ErrNoRecord
		}
		return 0, 0, err
	}
	return totalAmount, allocatedAmount, nil
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
	return item, nil
}

func (e ExpenseItemService) IsExist(expenseItemId int) (bool, error) {
	var exists bool
	stmt := `select exists (select 1 from expense_items where id = $1)`
	row := e.DB.QueryRow(stmt, expenseItemId)
	err := row.Scan(&exists)
	return exists, err
}

func (e ExpenseItemService) GetCreatorId(expenseItemId int) (int, error) {
	var creatorId int
	stmt := `select creator_id from expense_items where id = $1`
	row := e.DB.QueryRow(stmt, expenseItemId)
	err := row.Scan(&creatorId)
	return creatorId, err
}

func (e ExpenseItemService) Participate(
	allocatedAmount float32,
	updatedAt time.Time,
	expenseItemId int,
) error {
	stmt := `update expense_items set allocated_amount = allocated_amount + $1, last_updated_at = $2 where id = $3`
	_, err := e.DB.Exec(stmt, allocatedAmount, updatedAt, expenseItemId)
	return err
}

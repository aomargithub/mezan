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

func (e ExpenseItemService) GetExpenseId(expenseItemId int) (int, int, error) {
	stmt := " select mezani_id, expense_id from expense_items where id = $1"

	r := e.DB.QueryRow(stmt, expenseItemId)
	var mezaniId, expenseId int
	err := r.Scan(&mezaniId, &expenseId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, domain.ErrNoRecord
		}
		return 0, 0, err
	}
	return mezaniId, expenseId, nil
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

func (e ExpenseItemService) Participate(share domain.ExpenseItemShare) (float32, error) {
	var oldAmount float32
	stmt := `update expense_items set allocated_amount = $1, last_updated_at = $2 where id = $3`
	_, err := e.DB.Exec(stmt, share.Amount, share.CreatedAt, share.ExpenseItem.Id)
	if err != nil {
		return 0, err
	}

	stmt = `insert into expense_item_shares(created_at, share, amount, share_type, expense_item_id, expense_id, mezani_id,
                                participant_id)
				values ($1, $2, $3, $4, $5, $6, $7, $8)
				on conflict on constraint unique_participant_per_expense_item_share
					do update set share           = $2,
								  amount          = $3,
								  share_type      = $4,
								  last_updated_at = $1
				RETURNING (select old from expense_item_shares old where participant_id = $8 and expense_item_id = $5).amount`
	row := e.DB.QueryRow(stmt, share.CreatedAt, share.Share, share.Amount, share.ShareType, share.ExpenseItem.Id, share.Expense.Id,
		share.Mezani.Id, share.Participant.Id)

	err = row.Scan(&oldAmount)
	if err != nil {
		return 0, err
	}

	return oldAmount, nil
}

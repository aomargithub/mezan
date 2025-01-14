package db

import (
	"database/sql"
	"errors"
	"github.com/aomargithub/mezan/internal/domain"
)

type ExpenseService struct {
	DB *sql.DB
	dbCommons
}

func (e ExpenseService) Create(expense domain.Expense) error {
	stmt := `insert into expenses (name, total_amount, mezani_id, creator_id, created_at) values($1,$2,$3,$4,$5)`
	_, err := e.DB.Exec(stmt, expense.Name, expense.TotalAmount,
		expense.Mezani.Id, expense.Creator.Id, expense.CreatedAt)
	return err
}

func (e ExpenseService) Get(id int) (domain.Expense, error) {
	var expense domain.Expense
	stmt := `select e.id              as expense_id,
				   e.name            as expense_name,
				   e.created_at      as expense_created_at,
				   e.allocated_amount  as expense_allocated_amount,
				   e.total_amount    as expense_total_amount,
				   u1.name           as expense_creator_name,
				   COALESCE(ei.id, 0)             as item_id,
				   COALESCE(ei.name, '')          as item_name,
				   COALESCE(ei.allocated_amount, 0) as item_allocated_amount,
				   COALESCE(ei.amount, 0)         as item_amount,
				   COALESCE(ei.total_amount, 0)   as item_total_amount,
				   COALESCE(ei.quantity, 0)       as item_quantity
			from expenses e
					 join users u1 on u1.id = e.creator_id
					 left outer join expense_items ei on ei.expense_id = e.id
					 left outer join users u2 on ei.creator_id = u2.id
			where e.id = $1;`

	rows, err := e.DB.Query(stmt, id)
	if err != nil {
		return expense, err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.ExpenseItem
		err := rows.Scan(
			&expense.Id,
			&expense.Name,
			&expense.CreatedAt,
			&expense.AllocatedAmount,
			&expense.TotalAmount,
			&expense.Creator.Name,
			&item.Id,
			&item.Name,
			&item.AllocatedAmount,
			&item.Amount,
			&item.TotalAmount,
			&item.Quantity,
		)
		if err != nil {
			return domain.Expense{}, err
		}

		if item.Id != 0 {
			expense.Items = append(expense.Items, item)
		}
	}
	if err := rows.Err(); err != nil {
		return domain.Expense{}, err
	}
	if expense.Id == 0 {
		return domain.Expense{}, domain.ErrNoRecord
	}
	return expense, nil
}

func (e ExpenseService) GetMezaniId(expenseId int) (int, error) {
	stmt := " select mezani_id from expenses where id = $1"

	r := e.DB.QueryRow(stmt, expenseId)
	var mezaniId int
	err := r.Scan(&mezaniId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrNoRecord
		}
		return 0, err
	}
	return mezaniId, nil
}

func (e ExpenseService) IsExist(expenseId int) (bool, error) {
	var exists bool
	stmt := `select exists (select 1 from expenses where id = $1)`
	row := e.DB.QueryRow(stmt, expenseId)
	err := row.Scan(&exists)
	return exists, err
}

func (e ExpenseService) AddAmount(expenseId int, amount float32) error {
	stmt := `update expenses set total_amount=total_amount + $1 where id = $2`
	r, err := e.DB.Exec(stmt, amount, expenseId)
	if err != nil {
		return err
	}
	rowsAffected, err := r.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrNoRecord
	}
	return nil
}

func (e ExpenseService) Participate(share domain.ExpenseShare) (float32, error) {
	var oldAmount float32
	stmt := `update expenses set allocated_amount = allocated_amount + $1, last_updated_at = $2 where id = $3`
	_, err := e.DB.Exec(stmt, share.Amount, share.CreatedAt, share.Expense.Id)
	if err != nil {
		return 0, err
	}

	stmt = `insert into expense_shares(created_at, share, amount, share_type, expense_id, mezani_id, participant_id) 
			values ($1, $2, $3, $4, $5, $6, $7)
			on conflict on constraint unique_participant_per_expense_share 
			do update set amount = expense_shares.amount + $3,
			              share = ,
			              share_type = $4,
			              last_updated_at = $1
			RETURNING (select old from expense_shares old where participant_id = $7 and expense_id = $5).amount`
	row := e.DB.QueryRow(stmt, share.CreatedAt, share.Share, share.Amount, share.ShareType, share.Expense.Id,
		share.Mezani.Id, share.Participant.Id)
	err = row.Scan(&oldAmount)
	if err != nil {
		return 0, err
	}

	return oldAmount, nil
}

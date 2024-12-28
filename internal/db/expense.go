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
				   e.settled_amount  as expense_settled_amount,
				   e.total_amount    as expense_total_amount,
				   u1.name           as expense_creator_name,
				   COALESCE(ei.id, 0)             as item_id,
				   COALESCE(ei.name, '')          as item_name,
				   COALESCE(ei.settled_amount, 0) as item_settled_amount,
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
			&expense.SettledAmount,
			&expense.TotalAmount,
			&expense.Creator.Name,
			&item.Id,
			&item.Name,
			&item.SettledAmount,
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
		return domain.Expense{}, ErrNoRecord
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
			return 0, ErrNoRecord
		}
		return 0, err
	}
	return mezaniId, nil
}

func (e ExpenseService) Settle(expenseId int, amount float32) error {
	stmt := `update expenses set settled_amount = settled_amount + $1 where id = $2`
	_, err := e.DB.Exec(stmt, amount, expenseId)
	return err
}

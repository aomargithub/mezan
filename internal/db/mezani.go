package db

import (
	"database/sql"
	"github.com/aomargithub/mezan/internal/domain"
)

type MezaniService struct {
	DB *sql.DB
	commonDB
}

func (s MezaniService) Create(mezani domain.Mezani) (int, error) {
	stmt := `insert into mezanis (name, creator_id, share_id, created_at) values($1,$2,$3,$4) returning id`
	var id int
	row := s.DB.QueryRow(stmt, mezani.Name, mezani.Creator.Id, mezani.ShareId, mezani.CreatedAt)
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s MezaniService) Get(id int) (domain.Mezani, error) {
	var mezani domain.Mezani
	stmt := `select m.id                                                          as mezani_id,
				   m.name                                                        as mezani_name,
				   m.total_amount                                                   mezani_total_amount,
				   m.settled_amount                                                 mezani_settled_amount,
				   m.created_at                                                  as mezani_created_at,
				   u1.name                                                       as mezani_creator_name,
				   COALESCE(e.id, 0)                                             as expense_id,
				   COALESCE(e.name, '')                                          as expense_name,
				   COALESCE(e.total_amount, 0)                                   as expense_total_amount,
				   COALESCE(e.settled_amount, 0)                                 as expense_settled_amount,
				   COALESCE(e.created_at, '0001-01-01 00:00:00+00'::timestamptz) as expense_created_at,
				   COALESCE(u2.name, '')                                         as expense_creator_name
			from mezanis m
					 join users u1 on m.creator_id = u1.id
					 left outer join expenses e on e.mezani_id = m.id
					 left outer join users u2 on e.creator_id = u2.id
			where m.id = $1;`
	rows, err := s.DB.Query(stmt, id)
	if err != nil {
		return mezani, err
	}
	defer s.close(rows)
	for rows.Next() {
		var expense domain.Expense
		err = rows.Scan(&mezani.Id, &mezani.Name, &mezani.TotalAmount, &mezani.SettledAmount, &mezani.CreatedAt, &mezani.Creator.Name,
			&expense.Id, &expense.Name, &expense.TotalAmount, &expense.SettledAmount, &expense.CreatedAt, &expense.Creator.Name)
		if err != nil {
			return domain.Mezani{}, err
		}

		if expense.Id != 0 {
			mezani.Expenses = append(mezani.Expenses, expense)
		}
	}
	if err := rows.Err(); err != nil {
		return domain.Mezani{}, err
	}
	if mezani.Id == 0 {
		return domain.Mezani{}, ErrNoRecord
	}
	return mezani, nil
}

func (s MezaniService) GetAll() ([]domain.Mezani, error) {
	var mezanis []domain.Mezani
	stmt := `select id, name, created_at from mezanis`
	rows, err := s.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer s.close(rows)
	for rows.Next() {
		var mezani domain.Mezani
		err := rows.Scan(&mezani.Id, &mezani.Name, &mezani.CreatedAt)

		if err != nil {
			return nil, err
		}
		mezanis = append(mezanis, mezani)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return mezanis, nil
}

func (s MezaniService) AddExpense(mezaniId int, amount float32) error {
	stmt := `update mezanis set total_amount = total_amount + $1 where id = $2`
	_, err := s.DB.Exec(stmt, amount, mezaniId)
	return err
}

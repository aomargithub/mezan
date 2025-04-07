package db

import (
	"database/sql"
	"github.com/aomargithub/mezan/internal/domain"
	"time"
)

type MezaniService struct {
	DB *sql.DB
	dbCommons
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
				   m.total_amount                                                as mezani_total_amount,
				   m.allocated_amount                                            as mezani_allocated_amount,
				   m.created_at                                                  as mezani_created_at,
				   u1.name                                                       as mezani_creator_name,
				   COALESCE(e.id, 0)                                             as expense_id,
				   COALESCE(e.name, '')                                          as expense_name,
				   COALESCE(e.total_amount, 0)                                   as expense_total_amount,
				   COALESCE(e.allocated_amount, 0)                               as expense_allocated_amount,
				   COALESCE(e.has_items, false)                                  as expense_has_items,
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
		err = rows.Scan(&mezani.Id, &mezani.Name, &mezani.TotalAmount, &mezani.AllocatedAmount, &mezani.CreatedAt, &mezani.Creator.Name,
			&expense.Id, &expense.Name, &expense.TotalAmount, &expense.AllocatedAmount, &expense.HasItems, &expense.CreatedAt, &expense.Creator.Name)
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
		return domain.Mezani{}, domain.ErrNoRecord
	}
	return mezani, nil
}

func (s MezaniService) GetByShareId(shareId string) (domain.Mezani, error) {
	var mezani domain.Mezani
	stmt := `select m.id                                                          as mezani_id,
				   m.name                                                        as mezani_name,
				   m.total_amount                                                as mezani_total_amount,
				   m.allocated_amount                                            as mezani_allocated_amount,
				   m.created_at                                                  as mezani_created_at,
				   u1.name                                                       as mezani_creator_name,
				   COALESCE(e.id, 0)                                             as expense_id,
				   COALESCE(e.name, '')                                          as expense_name,
				   COALESCE(e.total_amount, 0)                                   as expense_total_amount,
				   COALESCE(e.allocated_amount, 0)                               as expense_allocated_amount,
				   COALESCE(e.created_at, '0001-01-01 00:00:00+00'::timestamptz) as expense_created_at,
				   COALESCE(u2.name, '')                                         as expense_creator_name
			from mezanis m
					 join users u1 on m.creator_id = u1.id
					 left outer join expenses e on e.mezani_id = m.id
					 left outer join users u2 on e.creator_id = u2.id
			where m.share_id = $1;`
	rows, err := s.DB.Query(stmt, shareId)
	if err != nil {
		return mezani, err
	}
	defer s.close(rows)
	for rows.Next() {
		var expense domain.Expense
		err = rows.Scan(&mezani.Id, &mezani.Name, &mezani.TotalAmount, &mezani.AllocatedAmount, &mezani.CreatedAt, &mezani.Creator.Name,
			&expense.Id, &expense.Name, &expense.TotalAmount, &expense.AllocatedAmount, &expense.CreatedAt, &expense.Creator.Name)
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
		return domain.Mezani{}, domain.ErrNoRecord
	}
	return mezani, nil
}

func (s MezaniService) GetAll(userId int) ([]domain.Mezani, error) {
	var mezanis []domain.Mezani
	stmt := `select m.id, m.name, m.created_at, m.share_id, m.total_amount, m.allocated_amount, u.name
				from mezanis m
						 join mezani_membership mm on mm.mezani_id = m.id
						 join users u on u.id = m.creator_id
				where mm.member_id = $1 order by m.created_at desc limit 10`
	rows, err := s.DB.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	defer s.close(rows)
	for rows.Next() {
		var mezani domain.Mezani
		err := rows.Scan(&mezani.Id, &mezani.Name, &mezani.CreatedAt, &mezani.ShareId, &mezani.TotalAmount, &mezani.AllocatedAmount, &mezani.Creator.Name)

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

func (s MezaniService) IsExist(mezaniId int) (bool, error) {
	var exists bool
	stmt := `select exists (select 1 from mezanis where id = $1)`
	row := s.DB.QueryRow(stmt, mezaniId)
	err := row.Scan(&exists)
	return exists, err
}

func (s MezaniService) AddAmount(mezaniId int, amount float32) error {
	stmt := `update mezanis set total_amount = total_amount + $1 where id = $2`
	result, err := s.DB.Exec(stmt, amount, mezaniId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNoRecord
	}
	return nil
}

func (e MezaniService) Participate(
	allocatedAmount float32,
	updatedAt time.Time,
	mezaniId int,
) error {
	stmt := `update mezanis set allocated_amount = allocated_amount + $1, last_updated_at = $2 where id = $3`
	_, err := e.DB.Exec(stmt, allocatedAmount, updatedAt, mezaniId)
	return err
}

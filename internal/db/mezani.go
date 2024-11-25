package db

import (
	"database/sql"
	"errors"

	"github.com/aomargithub/mezan/internal/domain"
)

type MezaniService struct {
	DB *sql.DB
}

func (s MezaniService) Create(mezani domain.Mezani) error {
	stmt := `insert into mezanis (name, creator_id, created_at) values($1,$2,$3)`
	_, err := s.DB.Exec(stmt, mezani.Name, mezani.CreatorId, mezani.CreatedAt)

	if err != nil {
		return err

	}
	return nil
}

func (s MezaniService) Get(id int) (domain.Mezani, error) {

	var mezani domain.Mezani
	stmt := `select id, name, created_at from mezanis where id = $1`
	row := s.DB.QueryRow(stmt, id)
	err := row.Scan(&mezani.Id, &mezani.Name, &mezani.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Mezani{}, ErrNoRecord
		}
		return domain.Mezani{}, err

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

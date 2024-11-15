package db

import (
	"database/sql"
	"errors"

	"github.com/aomargithub/mezan/internal/domain"
	"github.com/google/uuid"
)

type MezaniService struct {
	DB *sql.DB
}


func (service MezaniService) Create (mezani domain.Mezani) error  {
	stmt := `insert into mezanis (id, name, created_at) values($1,$2,$3)`
    _, err := service.DB.Exec(stmt, mezani.Id, mezani.Name, mezani.CreatedAt)
    
    if err != nil {
        return err
        
    }
    return nil
}


func (service MezaniService) Get (id uuid.UUID) (domain.Mezani, error) {

    var mezani domain.Mezani
    stmt := `select id, name, created_at from mezanis where id::text = $1`
    row := service.DB.QueryRow(stmt, id.String())
    err := row.Scan(&mezani.Id, &mezani.Name, &mezani.CreatedAt)

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return domain.Mezani{}, ErrNoRecord
        } else {
            return domain.Mezani{}, err
        }
    }
    return mezani, nil
}

func (service MezaniService) GetAll () ([]domain.Mezani, error) {

    var mezanis []domain.Mezani
    stmt := `select id, name, created_at from mezanis`
    rows, err := service.DB.Query(stmt)

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
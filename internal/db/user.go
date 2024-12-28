package db

import (
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"

	"github.com/aomargithub/mezan/internal/domain"
)

type UserService struct {
	DB *sql.DB
	dbCommons
}

func (s UserService) Create(user domain.User, hashedPassword string) error {
	stmt := `insert into users (name, email, hashed_password, created_at) values ($1, $2, $3, $4)`
	_, err := s.DB.Exec(stmt, user.Name, user.Email, hashedPassword, user.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && strings.Contains(pgErr.Message, "users_email_key") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (s UserService) GetInfoAndHashedPassword(email string) (domain.User, string, error) {
	stmt := `select id, name, hashed_password from users where email = $1`
	row := s.DB.QueryRow(stmt, email)
	var (
		user           domain.User
		hashedPassword string
	)
	err := row.Scan(&user.Id, &user.Name, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, "", ErrNoRecord
		}
		return domain.User{}, "", err
	}
	return user, hashedPassword, nil
}

func (s UserService) Exists(id int) (bool, error) {
	stmt := `select exists(select 1 from users where id  = $1)`

	var exists bool
	err := s.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}

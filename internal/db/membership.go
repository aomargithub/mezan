package db

import (
	"database/sql"
	"errors"
	"github.com/aomargithub/mezan/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"
)

type MembershipService struct {
	DB *sql.DB
}

func (s MembershipService) Create(membership domain.MemberShip) error {
	stmt := "insert into mezani_membership (created_at, member_id, mezani_id) values ($1,$2,$3)"
	_, err := s.DB.Exec(stmt, membership.CreatedAt, membership.Member.Id, membership.Mezani.Id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && strings.Contains(pgErr.Message, "unique_member_per_mezani") {
				return domain.ErrDuplicateRecord
			}
		}
		return err
	}
	return nil
}

func (s MembershipService) MezaniAccessibleBy(mezaniId int, userId int) (bool, error) {
	var allowed bool
	stmt := `select exists (select 1 from mezani_membership where mezani_id = $1 and member_id = $2)`
	row := s.DB.QueryRow(stmt, mezaniId, userId)
	err := row.Scan(&allowed)
	return allowed, err
}

func (s MembershipService) ExpenseAccessibleBy(expenseId int, userId int) (bool, error) {
	var allowed bool
	stmt := `select exists (select 1 from expenses e join mezani_membership m on e.mezani_id = m.mezani_id where e.id = $1 and m.member_id = $2)`
	row := s.DB.QueryRow(stmt, expenseId, userId)
	err := row.Scan(&allowed)
	return allowed, err
}

func (s MembershipService) ExpenseItemAccessibleBy(expenseItemId int, userId int) (bool, error) {
	var allowed bool
	stmt := `select exists (select 1 from expense_items e join mezani_membership m on e.mezani_id = m.mezani_id where e.id = $1 and m.member_id = $2)`
	row := s.DB.QueryRow(stmt, expenseItemId, userId)
	err := row.Scan(&allowed)
	return allowed, err
}

func (s MembershipService) PaymentAccessibleBy(paymentId int, userId int) (bool, error) {
	var allowed bool
	stmt := `select exists (select 1 from payments p join mezani_membership m on p.mezani_id = m.mezani_id where p.id = $1 and m.member_id = $2)`
	row := s.DB.QueryRow(stmt, paymentId, userId)
	err := row.Scan(&allowed)
	return allowed, err
}

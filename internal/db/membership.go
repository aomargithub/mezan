package db

import (
	"database/sql"
	"github.com/aomargithub/mezan/internal/domain"
)

type MembershipService struct {
	DB *sql.DB
}

func (s MembershipService) Create(membership domain.MemberShip) error {
	stmt := "insert into mezani_membership (created_at, member_id, mezani_id) values ($1,$2,$3)"
	_, err := s.DB.Exec(stmt, membership.CreatedAt, membership.Member.Id, membership.Mezani.Id)
	return err
}

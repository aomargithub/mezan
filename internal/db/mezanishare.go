package db

import (
	"database/sql"
	"github.com/aomargithub/mezan/internal/domain"
)

type MezaniShareService struct {
	DB *sql.DB
	dbCommons
}

func (e MezaniShareService) ParticipateInChild(share domain.MezaniShare) error {
	stmt := `insert into mezani_shares(created_at, share, amount, share_type, mezani_id, participant_id) 
			values ($1, $2, $3, $4, $5, $6)
			on conflict on constraint unique_participant_per_mezani_share 
			do update set amount = mezani_shares.amount + $3,
			              share = mezani_shares.share + $2,
			              share_type = $4,
			              last_updated_at = $1`
	_, err := e.DB.Exec(stmt, share.CreatedAt, share.Share, share.Amount, share.ShareType,
		share.Mezani.Id, share.Participant.Id)
	return err
}

func (e MezaniShareService) Participate(share domain.MezaniShare) error {
	stmt := `insert into mezani_shares(created_at, share, amount, share_type, mezani_id, participant_id) 
			values ($1, $2, $3, $4, $5, $6)
			on conflict on constraint unique_participant_per_mezani_share 
			do update set amount = $3,
			              share = $2,
			              share_type = $4,
			              last_updated_at = $1`
	_, err := e.DB.Exec(stmt, share.CreatedAt, share.Share, share.Amount, share.ShareType,
		share.Mezani.Id, share.Participant.Id)
	return err
}

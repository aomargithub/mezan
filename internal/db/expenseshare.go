package db

import (
	"database/sql"
	"github.com/aomargithub/mezan/internal/domain"
)

type ExpenseShareService struct {
	DB *sql.DB
	dbCommons
}

func (e ExpenseShareService) ParticipateInItem(share domain.ExpenseShare) error {
	stmt := `insert into expense_shares(created_at, share, amount, share_type, expense_id, mezani_id, participant_id) 
			values ($1, $2, $3, $4, $5, $6, $7)
			on conflict on constraint unique_participant_per_expense_share 
			do update set amount = expense_shares.amount + $3,
			              share = expense_shares.share + $2,
			              share_type = $4,
			              last_updated_at = $1`
	_, err := e.DB.Exec(stmt, share.CreatedAt, share.Share, share.Amount, share.ShareType, share.Expense.Id,
		share.Mezani.Id, share.Participant.Id)
	return err
}

func (e ExpenseShareService) Participate(share domain.ExpenseShare) (float32, error) {
	var oldAmount float32
	stmt := `insert into expense_shares(created_at, share, amount, share_type, expense_id, mezani_id, participant_id) 
			values ($1, $2, $3, $4, $5, $6, $7)
			on conflict on constraint unique_participant_per_expense_share 
			do update set amount = expense_shares.amount + $3,
			              share = expense_shares.share + $2,
			              share_type = $4,
			              last_updated_at = $1
			RETURNING (select old from expense_shares old where participant_id = $7 and expense_id = $5).amount`
	row := e.DB.QueryRow(stmt, share.CreatedAt, share.Share, share.Amount, share.ShareType, share.Expense.Id,
		share.Mezani.Id, share.Participant.Id)
	err := row.Scan(&oldAmount)
	if err != nil {
		return 0, err
	}

	return oldAmount, nil
}

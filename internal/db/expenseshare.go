package db

import (
	"database/sql"
	"errors"
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

func (e ExpenseShareService) Participate(share domain.ExpenseShare) (*float32, error) {
	var oldAmount *float32
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
	err := row.Scan(oldAmount)
	if err != nil {
		return nil, err
	}
	return oldAmount, nil
}

func (e ExpenseShareService) GetByExpenseIdParticipantId(
	expenseId int,
	participantId int,
) (domain.ExpenseShare, error) {
	var expenseShare domain.ExpenseShare
	stmt := `select share_type, share, amount from expense_shares where expense_id = $1 and participant_id = $2`
	row := e.DB.QueryRow(stmt, expenseId, participantId)
	err := row.Scan(&expenseShare.ShareType, &expenseShare.Share, &expenseShare.Amount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ExpenseShare{}, domain.ErrNoRecord
		}
		return domain.ExpenseShare{}, err
	}

	return expenseShare, nil
}

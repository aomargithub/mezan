package db

import (
	"database/sql"
	"github.com/aomargithub/mezan/internal/domain"
)

type ExpenseItemShareService struct {
	DB *sql.DB
	dbCommons
}

func (e ExpenseItemShareService) Participate(share domain.ExpenseItemShare) (*float32, error) {
	var oldAmount *float32
	stmt := `insert into expense_item_shares(created_at, share, amount, share_type, expense_item_id, expense_id, mezani_id,
                                participant_id)
				values ($1, $2, $3, $4, $5, $6, $7, $8)
				on conflict on constraint unique_participant_per_expense_item_share
					do update set share           = $2,
								  amount          = $3,
								  share_type      = $4,
								  last_updated_at = $1
				RETURNING (select old from expense_item_shares old where participant_id = $8 and expense_item_id = $5).amount`
	row := e.DB.QueryRow(stmt, share.CreatedAt, share.Share, share.Amount, share.ShareType, share.ExpenseItem.Id, share.Expense.Id,
		share.Mezani.Id, share.Participant.Id)
	err := row.Scan(&oldAmount)
	if err != nil {
		return nil, err
	}
	return oldAmount, nil
}

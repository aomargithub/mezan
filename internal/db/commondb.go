package db

import "database/sql"

type dbCommons struct {
}

func (c dbCommons) close(r *sql.Rows) {
	_ = r.Close()
}

func (c dbCommons) Rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}

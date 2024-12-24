package db

import "database/sql"

type commonDB struct {
}

func (s commonDB) close(r *sql.Rows) {
	_ = r.Close()
}

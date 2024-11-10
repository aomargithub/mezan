package db

import (
	"time"
    "fmt"
    "database/sql"
	"github.com/aomargithub/mezan/internal/domain"
	"github.com/google/uuid"
)

type MezanaRepo struct {
	DB *sql.DB
}


func (repo MezanaRepo) Insert (mezana domain.Mezana) (int, error)  {

	stmt := `INSERT INTO mezanas (id, name, created_at) VALUES($1,$2,$3)`

    // Use the Exec() method on the embedded connection pool to execute the
    // statement. The first parameter is the SQL statement, followed by the
    // values for the placeholder parameters: title, content and expiry in
    // that order. This method returns a sql.Result type, which contains some
    // basic information about what happened when the statement was executed.
    result, err := repo.DB.Exec(stmt, uuid.New(), mezana.Name, time.Now())
    fmt.Println(stmt, result, err)
    if err != nil {
        return 0, err
    }

    // Use the LastInsertId() method on the result to get the ID of our
    // newly inserted record in the snippets table.
    id, err := result.LastInsertId()
    if err != nil {
        return 0, err
    }

    // The ID returned has the type int64, so we convert it to an int type
    // before returning.
    return int(id), nil
}
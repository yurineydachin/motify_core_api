package interfaces

import (
	"database/sql"
)

// IQueryer interface for sqlx.Tx and sql.DB
type IQueryer interface {
	Select(dest interface{}, query string, args ...interface{}) error
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
}

package sqlbatch

import (
	"fmt"

	"database/sql"
)

// Command format for sending a batch of sql commands.
// Query is the sql query to execute (required).
// ArgsFunc is called before execution for query arguments (optional).
// Args are query parameters (optional). Ignored if ArgsFunc is non-nil.
// ScanOnce is the scan function for reading at most one row (optional).
// Scan is the scan function for reading each row (optional).
// If ScanOnce is non-nil, Scan is ignored.
// Affect is the number of rows that should be affected.
// If Affect is zero (default), it is not checked.
// If Affect is negative, no rows should be affected.
// If Affect is positive, that should be the number of affected rows.
type Command struct {
	Query    string
	ArgsFunc func() []interface{}
	Args     []interface{}
	ScanOnce func(fn func(...interface{}) error) error
	Scan     func(fn func(...interface{}) error) error
	Affect   int64
}

// Handler contains the database handle.
type Handler struct {
	db *sql.DB
}

// New creates a new handler for handling batch SQL operations.
func New(db *sql.DB) Handler {
	return Handler{db: db}
}

// Batch executes a batch of commands in a single transaction.
// If any error occurs, the transaction will be rolled back.
func (handler Handler) Batch(commands []Command) error {
	tx, err := handler.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, command := range commands {
		args := command.Args
		if command.ArgsFunc != nil {
			args = command.ArgsFunc()
		}
		if command.Affect != 0 {
			result, err := tx.Exec(command.Query, args...)
			if err != nil {
				return err
			}
			affected, err := result.RowsAffected()
			if err != nil {
				return err
			}
			expected := command.Affect
			if expected < 0 {
				expected = 0
			}
			if expected != affected {
				err = fmt.Errorf(expectedDifferentAffectedRows, expected, affected, command.Query)
				return err
			}
		} else {
			rows, err := tx.Query(command.Query, args...)
			if err != nil {
				return err
			}
			if command.ScanOnce != nil {
				if rows.Next() {
					err = command.ScanOnce(rows.Scan)
					if err != nil {
						return err
					}
				}
			} else if command.Scan != nil {
				for rows.Next() {
					err = command.Scan(rows.Scan)
					if err != nil {
						return err
					}
				}
			}
			if err = rows.Err(); err != nil {
				rows.Close()
				return err
			}
			err = rows.Close()
			if err != nil {
				return err
			}
		}
	}

	tx.Commit()
	return nil
}

const expectedDifferentAffectedRows = "Expected to affect %v rows, but %v rows affected for query: `%v`"

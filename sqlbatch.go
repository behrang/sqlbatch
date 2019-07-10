package sqlbatch

import (
	"fmt"

	"database/sql"
)

// Command format for sending a batch of sql commands.
// Query is the sql query to execute (required).
// ArgsFunc is called before execution for query arguments (optional).
// It will be passed current results.
// Args are query parameters (optional). Ignored if ArgsFunc is non-nil.
// Init is the default value passed to first iteration of ReadAll.
// Following iterations will use the previous memo returned by ReadAll.
// ReadAll is the read function for reading all rows (optional).
// ReadOne is the read function for reading at most one row (optional).
// If ReadOne is non-nil, ReadAll is ignored.
// Affect is the number of rows that should be affected.
// If Affect is zero (default), it is not checked.
// If Affect is negative, no rows should be affected.
// If Affect is positive, that should be the number of affected rows.
type Command struct {
	Query    string
	ArgsFunc func([]interface{}) []interface{}
	Args     []interface{}
	Init     interface{}
	ReadAll  func(memo interface{}, fn func(...interface{}) error) (interface{}, error)
	ReadOne  func(fn func(...interface{}) error) (interface{}, error)
	Affect   int64
}

// Batch executes a batch of commands in a single transaction.
// It will return a results, and an error. Results will include
// the result returned by ReadAll or ReadOne for each command
// at the specific index.
func Batch(tx *sql.Tx, commands []Command) ([]interface{}, error) {

	results := make([]interface{}, len(commands))
	for i, command := range commands {
		args := command.Args
		if command.ArgsFunc != nil {
			args = command.ArgsFunc(results)
		}
		if command.Affect != 0 {
			result, err := tx.Exec(command.Query, args...)
			if err != nil {
				return results, err
			}
			affected, err := result.RowsAffected()
			if err != nil {
				return results, err
			}
			expected := command.Affect
			if expected < 0 {
				expected = 0
			}
			if expected != affected {
				err = fmt.Errorf(expectedDifferentAffectedRows, expected, affected, command.Query)
				return results, err
			}
		} else {
			rows, err := tx.Query(command.Query, args...)
			if err != nil {
				return results, err
			}
			defer rows.Close()
			if command.ReadOne != nil {
				if rows.Next() {
					result, err := command.ReadOne(rows.Scan)
					if err != nil {
						return results, err
					}
					results[i] = result
				}
			} else if command.ReadAll != nil {
				memo := command.Init
				for rows.Next() {
					memo, err = command.ReadAll(memo, rows.Scan)
					if err != nil {
						return results, err
					}
				}
				results[i] = memo
			}
			if err = rows.Err(); err != nil {
				return results, err
			}
			err = rows.Close()
			if err != nil {
				return results, err
			}
		}
	}
	return results, nil
}

const expectedDifferentAffectedRows = "Expected to affect %v rows, but %v rows affected for query: `%v`"

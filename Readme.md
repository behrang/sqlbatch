# SQLBatch

Executes a batch of SQL commands in a single transaction.

## Install

```bash
go get -u github.com/behrang/sqlbatch
```

## Usage

First create a connection pool for your database, for example:

```go
db := sql.Open("postgres", "connectionInfo")
```

Then create a batch handler:

```go
batchHandler := sqlbatch.New(db)
```

Then to execute a batch of commands:

```go
var dynamicParam int
people := make([]Person, 0)
err := batchHandler.Batch([]Command{
    {
        Query: `A SQL query`,
        Scan: func(scan func(...interface{}) error) error {
            return scan(&dynamicParam)
        },
    },
    {
        Query: `Another SQL query with $1 and $2 parameters`,
        Args: []interface{}{param1, param2},
        Affect: 1,
    },
    {
        Query: `Yet another SQL query to scan results with a dynamic param $1`,
        ArgsFunc: func() []interface{} {
            return []interface{}{dynamicParam}
        },
        Scan: func(scan func(...interface{}) error) error {
            p := Person{}
            err := scan(&p.Name, &p.Age)
            people = append(people, person)
            return err
        },
    }
})
```

If any of the queries fail, transaction will be rolled back.

`Affect` checks the affected rows. If it is a positive number, the number of affected rows should be equal to it or transaction will be rolled back. If is a negative number, no rows should be affected or transaction will be rolled back. If it is 0 or omitted, affected rows will not be checked.

`Args` provides arguments to the SQL prepared statement. These args should be final at the time of batch command creation.

`ArgsFunc` provides dynamic parameters to the SQL prepared statement. If args for a query need to be calculated from other queries in the same batch, use this function.

`Scan` scans each row of the results. If an error is returned, transaction will be rolled back.

`ScanOnce` scans at most one row. More rows will be ignored. If no row exists, nothing will be scanned. If an error is returned, transaction will be rolled back.

If no error occurs, the transaction will be committed. If an error occurs while commiting, the error is returned. Intermediary rows and result sets are cleaned up.

## License

MIT

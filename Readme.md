# SQLBatch

Executes a batch of SQL commands in a single transaction.

## Install

```bash
go get -u github.com/behrang/sqlbatch
```

## Usage

```go
db := sql.Open("postgres", "connectionInfo")
tx := db.BeginTx(context.Background(), nil)
results, err := sqlbatch.Batch(tx, []Command{
    {
        Query: `A SQL query`,
        ReadOne: func(scan func(...interface{}) error) (interface{}, error) {
            var x int
            err := scan(&x)
            return x, err
        },
    },
    {
        Query: `Another SQL query with $1 and $2 parameters`,
        Args: []interface{}{param1, param2},
        Affect: 1,
    },
    {
        Query: `Yet another SQL query to scan results with a dynamic param $1`,
        ArgsFunc: func(results []interface{}) []interface{} {
            return []interface{}{results[0].(int)}
        },
        Init: make([]Person, 0),
        ReadAll: func(memo interface{}, scan func(...interface{}) error) (interface{}, error) {
            p := Person{}
            err := scan(&p.Name, &p.Age)
            people := memo.([]Person)
            people = append(people, person)
            return people, err
        },
    }
})
if err != nil {
    tx.Rollback()
} else {
    tx.Commit()
    x, _ := results[0].(int)
    people, _ := results[2].([]Person)
    fmt.Println(x, people)
}
```

If any of the queries fail, transaction will be rolled back.

`Affect` checks the affected rows. If it is a positive number, the number of affected rows should be equal to it or transaction will be rolled back. If is a negative number, no rows should be affected or transaction will be rolled back. If it is 0 or omitted, affected rows will not be checked.

`Args` provides arguments to the SQL prepared statement. These args should be final at the time of batch command creation.

`ArgsFunc` provides dynamic parameters to the SQL prepared statement. If args for a query need to be calculated from results of other queries in the same batch, use this function.

`ReadAll` scans each row of the results. If an error is returned, transaction will be rolled back. A `memo` object can be used to add all results together, for example in an array or map. `Init` is the default value passed to first iteration of `ReadAll`.

`ReadOne` scans at most one row. More rows will be ignored. If no row exists, nothing will be scanned. If an error is returned, transaction will be rolled back.

If no error occurs, the transaction will be committed. If an error occurs while commiting, the error is returned. Intermediary rows and result sets are cleaned up.

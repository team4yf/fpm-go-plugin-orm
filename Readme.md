# fpm-go-plugin-orm

Plugin for fpm to access RDMS.

Support `mysql`, `postgres` .

## Install

get from the github.

`$ go get -u github.com/team4yf/fpm-go-plugin-orm`

import on the code file header.

```golang
import (
	"github.com/team4yf/fpm-go-plugin-orm/plugins"
)
```

## AutoMigration

All sql files under `migrations` will be automatically migrated.

```
- migrations
  - V1.2022.01.01.00__test.sql
```

`V1`: version of db, could be `V*`, keep `*` as increasingly number.
`2022.01.01`: date of migration
`00`: sequence of migration, reset when another day starts.
`test`: desc of migration

double underscore should be added between date and desc.

program will execute all migrations as the sort of migration name.


## Config

```json
{
    "db": {
        "engine": "postgres",
        "user": "postgres",
        "password": "root",
        "host": "localhost",
        "port": 5432,
        "database": "pg",
        "charset": "utf8",
        "showSql": true
    }
}
```

## Usage

```golang

dbclient, exists := app.GetDatabase("pg")
one := &Fake{}
q := db.NewQuery()
q.AddSorter(db.Sorter{
    Sortby: "name",
    Asc:    "asc",
}).SetTable("fake").SetCondition("name = ?", "c")
err = dbclient.First(q, &one)

```



## ChangeLog

v0.0.2
Feature:
- Support `Select(string)` to define the fetch columens.
- Add common bizModel.
- Change the query function


v0.0.1
Feature:
- Basic CURD 
- Pagination
- Transaction
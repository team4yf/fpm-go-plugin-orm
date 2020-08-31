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
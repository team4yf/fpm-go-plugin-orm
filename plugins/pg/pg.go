package pg

import (
	"github.com/team4yf/fpm-go-plugin-orm/plugins"
	"github.com/team4yf/yf-fpm-server-go/fpm"
	"github.com/team4yf/yf-fpm-server-go/pkg/db"
	//import the postgress
)

type queryReq struct {
	Table     string `json:"table,omitempty"`
	Condition string `json:"condition,omitempty"`
	Skip      int    `json:"skip,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	ID        int64  `json:"id,omitempty"`
	Sort      string `json:"sort,omitempty"`
}

type dataRow map[string]interface{}

func init() {
	fpm.Register(func(app *fpm.Fpm) {
		option := &plugins.DBSetting{}
		if err := app.FetchConfig("db", &option); err != nil {
			panic(err)
		}
		dbInstance := plugins.New(option)
		dbclient := plugins.NewImpl(dbInstance)
		app.SetDatabase("pg", func() db.Database {
			return dbclient
		})
		bizModule := make(fpm.BizModule, 0)

		// support:
		// 1. 'find', 'first', 'create', 'update', 'remove', 'clear', 'get', 'count', 'findAndCount'

		bizModule["find"] = func(param *fpm.BizParam) (data interface{}, err error) {
			queryReq := queryReq{}
			err = param.Convert(&queryReq)
			if err != nil {
				return
			}

			q := db.NewQuery()
			q.SetTable(queryReq.Table)
			q.SetPager(&db.Pagination{
				Skip:  queryReq.Skip,
				Limit: queryReq.Limit,
			})
			list := make([]*dataRow, 0)
			err = dbclient.Find(q, &list)
			data = &list
			return
		}

		bizModule["remove"] = func(param *fpm.BizParam) (data interface{}, err error) {
			queryReq := queryReq{}
			err = param.Convert(&queryReq)
			if err != nil {
				return
			}

			q := db.NewQuery()
			q.SetTable(queryReq.Table)
			q.SetCondition(queryReq.Condition)
			var rows int64
			err = dbclient.Remove(q.BaseData, &rows)
			data = rows
			return
		}

		app.AddBizModule("common", &bizModule)
	})
}

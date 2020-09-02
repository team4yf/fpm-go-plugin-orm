package pg

import (
	"github.com/team4yf/fpm-go-plugin-orm/plugins"
	"github.com/team4yf/yf-fpm-server-go/fpm"
	"github.com/team4yf/yf-fpm-server-go/pkg/db"
)

type queryReq struct {
	Table     string      `json:"table,omitempty"`
	Condition string      `json:"condition,omitempty"`
	Fields    string      `json:"fields,omitempty"`
	Skip      int         `json:"skip,omitempty"`
	Limit     int         `json:"limit,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	ID        int64       `json:"id,omitempty"`
	Sort      string      `json:"sort,omitempty"`
}

func parseQuery(req *queryReq) *db.QueryData {
	q := db.NewQuery()
	q.SetTable(req.Table)
	if req.Limit != 0 {
		q.SetPager(&db.Pagination{
			Skip:  req.Skip,
			Limit: req.Limit,
		})
	}

	if req.Fields != "" {
		//TODO change the data  type
		// q.AddFields(strings.Split(queryReq.Fields, ","))
	}
	q.SetCondition(req.Condition)

	if req.Sort != "" {

		l := len(req.Sort)
		asc := "asc"
		if req.Sort[l-1:] == "-" {
			asc = "desc"
		}
		q.AddSorter(db.Sorter{
			Sortby: req.Sort[0 : l-1],
			Asc:    asc,
		})
	}

	return q
}
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
		// 1. x 'find', x 'first', 'create', 'update', x 'remove', x 'clear', x 'get', x 'count', x 'findAndCount'

		//TODO: sorter
		bizModule["find"] = func(param *fpm.BizParam) (data interface{}, err error) {
			queryReq := queryReq{}
			err = param.Convert(&queryReq)
			if err != nil {
				return
			}

			q := parseQuery(&queryReq)
			list := make([]map[string]interface{}, 0)
			err = dbclient.Find(q, &list)
			data = &list
			return
		}

		bizModule["findAndCount"] = func(param *fpm.BizParam) (data interface{}, err error) {
			queryReq := queryReq{}
			err = param.Convert(&queryReq)
			if err != nil {
				return
			}

			q := parseQuery(&queryReq)
			list := make([]map[string]interface{}, 0)
			var total int64
			err = dbclient.FindAndCount(q, &list, &total)

			data = map[string]interface{}{
				"total": total,
				"rows":  list,
			}
			return
		}

		bizModule["first"] = func(param *fpm.BizParam) (data interface{}, err error) {
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
			q.SetCondition(queryReq.Condition)
			one := make(map[string]interface{})
			err = dbclient.First(q, &one)
			data = &one
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

		bizModule["create"] = func(param *fpm.BizParam) (data interface{}, err error) {
			queryReq := queryReq{}
			err = param.Convert(&queryReq)
			if err != nil {
				return
			}

			q := db.NewQuery()
			q.SetTable(queryReq.Table)
			err = dbclient.Create(q.BaseData, queryReq.Data)
			data = 1
			return
		}

		app.AddBizModule("common", &bizModule)
	})
}

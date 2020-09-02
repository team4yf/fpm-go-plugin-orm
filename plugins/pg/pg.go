package pg

import (
	"strings"

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
	Data      interface{} `json:"row,omitempty"`
	ID        int64       `json:"id,omitempty"`
	Sort      string      `json:"sort,omitempty"`
}

func parseQueryFromBizParam(param *fpm.BizParam) (q *db.QueryData, err error) {
	queryReq := queryReq{}
	if err = param.Convert(&queryReq); err != nil {
		return
	}

	q = parseQuery(&queryReq)
	return
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
		q.AddFields((strings.Split(req.Fields, ","))...)
	}
	if req.Condition != "" {
		q.SetCondition(req.Condition)
	}
	if req.ID != 0 {
		q.SetCondition("id = ?", req.ID)
	}

	if req.Sort != "" {

		l := len(req.Sort)
		asc := "asc"
		sortBy := "id"
		lastLetter := req.Sort[l-1:]
		if lastLetter == "-" {
			asc = "desc"
		}

		if lastLetter != "-" && lastLetter != "+" {
			sortBy = req.Sort
		} else {
			sortBy = req.Sort[0 : l-1]
		}
		q.AddSorter(db.Sorter{
			Sortby: sortBy,
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

		bizModule["find"] = func(param *fpm.BizParam) (data interface{}, err error) {
			q, err := parseQueryFromBizParam(param)
			if err != nil {
				return nil, err
			}
			list := make([]map[string]interface{}, 0)
			err = dbclient.Find(q, &list)
			data = &list
			return
		}

		bizModule["findAndCount"] = func(param *fpm.BizParam) (data interface{}, err error) {
			q, err := parseQueryFromBizParam(param)
			if err != nil {
				return nil, err
			}
			list := make([]map[string]interface{}, 0)
			var total int64
			err = dbclient.FindAndCount(q, &list, &total)

			data = map[string]interface{}{
				"count": total,
				"rows":  list,
			}
			return
		}

		bizModule["count"] = func(param *fpm.BizParam) (data interface{}, err error) {
			q, err := parseQueryFromBizParam(param)
			if err != nil {
				return nil, err
			}
			var total int64
			err = dbclient.Count(q.BaseData, &total)
			data = total
			return
		}

		bizModule["first"] = func(param *fpm.BizParam) (data interface{}, err error) {
			q, err := parseQueryFromBizParam(param)
			if err != nil {
				return nil, err
			}
			one := make(map[string]interface{})
			err = dbclient.First(q, &one)
			data = &one
			return
		}

		bizModule["get"] = func(param *fpm.BizParam) (data interface{}, err error) {
			req := queryReq{}
			if err = param.Convert(&req); err != nil {
				return
			}
			q := parseQuery(&req)
			q.SetCondition("id = ?", req.ID)
			one := make(map[string]interface{})
			err = dbclient.First(q, &one)
			data = &one
			return
		}

		bizModule["remove"] = func(param *fpm.BizParam) (data interface{}, err error) {
			req := queryReq{}
			if err = param.Convert(&req); err != nil {
				return
			}

			q := parseQuery(&req)
			q.SetCondition("id = ?", req.ID)
			var rows int64
			err = dbclient.Remove(q.BaseData, &rows)
			data = rows
			return
		}

		bizModule["clear"] = func(param *fpm.BizParam) (data interface{}, err error) {
			q, err := parseQueryFromBizParam(param)
			if err != nil {
				return nil, err
			}
			var rows int64
			err = dbclient.Remove(q.BaseData, &rows)
			data = rows
			return
		}

		bizModule["create"] = func(param *fpm.BizParam) (data interface{}, err error) {
			req := queryReq{}
			if err = param.Convert(&req); err != nil {
				return
			}

			q := parseQuery(&req)
			q.SetTable(req.Table)
			err = dbclient.Create(q.BaseData, req.Data)
			data = 1
			return
		}

		bizModule["update"] = func(param *fpm.BizParam) (data interface{}, err error) {
			req := queryReq{}
			if err = param.Convert(&req); err != nil {
				return
			}

			q := parseQuery(&req)
			q.SetTable(req.Table)
			q.SetCondition(req.Condition)
			var rows int64
			cm := db.CommonMap{}
			for k, v := range (req.Data).(map[string]interface{}) {
				cm[k] = v
			}
			err = dbclient.Updates(q.BaseData, cm, &rows)
			data = rows
			return
		}

		app.AddBizModule("common", &bizModule)
	})
}

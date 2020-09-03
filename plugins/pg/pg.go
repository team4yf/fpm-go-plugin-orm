package pg

import (
	"fmt"
	"strings"

	"github.com/team4yf/yf-fpm-server-go/pkg/utils"

	"github.com/team4yf/fpm-go-plugin-orm/plugins"
	"github.com/team4yf/yf-fpm-server-go/fpm"
	"github.com/team4yf/yf-fpm-server-go/pkg/db"
)

type queryReq struct {
	Table     string      `json:"table,omitempty"`
	Condition interface{} `json:"condition,omitempty"`
	Fields    string      `json:"fields,omitempty"`
	Skip      int         `json:"skip,omitempty"`
	Limit     int         `json:"limit,omitempty"`
	Data      interface{} `json:"row,omitempty"`
	ID        interface{} `json:"id,omitempty"`
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
		//TODO： 这里尚未考虑到兼容性
		f := strings.ReplaceAll(req.Fields, "updateAt", "updated_at,(floor(extract(epoch from updated_at) *1000)::bigint) as updateAt")
		f = strings.ReplaceAll(f, "createAt", "created_at,(floor(extract(epoch from created_at) *1000)::bigint) as createAt")
		q.AddFields((strings.Split(f, ","))...)
	}
	if req.Condition != nil {
		switch req.Condition.(type) {
		case string:
			q.SetCondition(req.Condition.(string))
		case map[string]interface{}:
			conditions := req.Condition.(map[string]interface{})
			keys := make([]string, 0)
			vals := make([]interface{}, 0)
			for k, v := range conditions {
				keys = append(keys, k+" = ?")
				vals = append(vals, v)
			}
			q.SetCondition(strings.Join(keys, ","), vals...)
		default:
			fmt.Printf("what? %v\n", req.Condition)
		}

	}
	if req.ID != nil {
		//对ID的类型进行判断
		switch req.ID.(type) {
		case float64:
			q.SetCondition("id = ?", (int64)(req.ID.(float64)))
		case int64:
			q.SetCondition("id = ?", req.ID.(int64))
		default:
			q.SetCondition("id = ?", req.ID)
		}

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
			//here, it's unsafe, the condition could be interface{}
			// q.SetCondition(req.Condition.(string))
			var rows int64
			cm := db.CommonMap{}
			if err = utils.Interface2Struct(req.Data, &cm); err != nil {
				return
			}
			err = dbclient.Updates(q.BaseData, cm, &rows)
			data = rows
			return
		}

		app.AddBizModule("common", &bizModule)
	})
}

package fake

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/team4yf/yf-fpm-server-go/fpm"
	"github.com/team4yf/yf-fpm-server-go/pkg/db"

	_ "github.com/team4yf/fpm-go-plugin-orm/plugins/pg"
)

type countBody struct {
	C float64 `json:"c"`
	B float64 `json:"b"`
}

func TestPG(t *testing.T) {
	app := fpm.New()

	app.Init()

	dbclient, exists := app.GetDatabase("pg")
	assert.Equal(t, true, exists, "should true")

	//Test AutoMigrate
	err := dbclient.AutoMigrate(
		&Fake{},
	)

	assert.Nil(t, err, "should nil err")

	var rows, total int64
	//Test Execute
	rows = 0
	err = dbclient.Execute(`delete from fake where name = 'c'`, &rows)
	assert.Nil(t, err, "should not error")
	assert.Equal(t, true, rows >= 0, "should gt 0")

	//Test Remove
	rows = 0
	q := db.NewQuery()
	q.SetTable("fake").SetCondition("name = ?", "b")
	err = dbclient.Remove(q.BaseData, &rows)
	assert.Nil(t, err, "should nil err")
	assert.Equal(t, true, rows >= 0, "should gt 0")

	//Test Create
	err = dbclient.Create(nil, &Fake{
		Name:  "c",
		Value: 100,
	})

	assert.Nil(t, err, "should nil err")

	//Test First
	one := &Fake{}
	q = db.NewQuery()
	q.AddSorter(db.Sorter{
		Sortby: "name",
		Asc:    "asc",
	}).SetTable("fake").SetCondition("name = ?", "c")
	err = dbclient.First(q, &one)

	assert.Equal(t, 100, one.Value, "should be 100")

	list := make([]*Fake, 0)
	q.AddSorter(db.Sorter{
		Sortby: "id",
		Asc:    "desc",
	})
	err = dbclient.Find(q, &list)

	assert.Equal(t, true, len(list) > 0, "should more data")

	//Test Count
	total = 0
	err = dbclient.Count(q.BaseData, &total)
	assert.Nil(t, err, "should nil err")
	assert.Equal(t, true, total > 0, "should gt 0")

	//Test Find&Count
	list = make([]*Fake, 0)
	err = dbclient.FindAndCount(q, &list, &total)
	assert.Nil(t, err, "should nil err")
	assert.Equal(t, true, len(list) > 0, "should gt 0")
	assert.Equal(t, true, total > 0, "should gt 0")

	//Test Raw
	raw := &countBody{}
	err = dbclient.Raw(`select count(1) as c, 1 as b from fake`, raw)
	assert.Nil(t, err, "should not error")
	assert.Equal(t, true, raw.C >= 0, "should gt 0")
	assert.Equal(t, true, raw.B == 1, "should eq 1")

	//Test Raws
	raws := make([]*countBody, 0)
	err = dbclient.Raws(`select id as c, 1 as b from fake`, func() interface{} {
		return &countBody{}
	}, func(one interface{}) {
		raws = append(raws, one.(*countBody))
	})
	assert.Nil(t, err, "should not error")
	assert.Equal(t, true, len(raws) >= 0, "should gt 0")

	//Test Updates
	fields := db.CommonMap{
		"value": 101,
		"name":  "b",
	}
	q.SetCondition("name=?", "c")
	err = dbclient.Updates(q.BaseData, fields, &total)
	assert.Nil(t, err, "should nil err")
	assert.Equal(t, true, total > 0, "should gt 0")

}

func TestTran(t *testing.T) {
	app := fpm.New()

	app.Init()

	dbclient, exists := app.GetDatabase("pg")
	assert.Equal(t, true, exists, "should true")
	//OK
	err := dbclient.Transaction(func(tx db.Database) (err error) {
		var total int64
		fields := db.CommonMap{
			"value": 101,
		}
		q := db.NewQuery()
		q.SetTable("fake").SetCondition("name = ?", "c")
		err = tx.Updates(q.BaseData, fields, &total)

		return
	})
	assert.Nil(t, err, "should nil err")
	//Fail
	err = dbclient.Transaction(func(tx db.Database) (err error) {
		var total int64
		fields := db.CommonMap{
			"value": 102,
		}
		q := db.NewQuery()
		q.SetTable("fake").SetCondition("name = ?", "c")
		err = tx.Updates(q.BaseData, fields, &total)

		return errors.New("err")
	})

	assert.NotNil(t, err, "should err")
}

func TestBiz(t *testing.T) {
	app := fpm.New()

	app.Init()
	dbclient, exists := app.GetDatabase("pg")
	assert.Equal(t, true, exists, "should true")

	//Test AutoMigrate
	err := dbclient.AutoMigrate(
		&Fake{},
	)

	assert.Nil(t, err, "should nil err")

	//Test Execute
	var rows int64
	err = dbclient.Execute(`delete from fake where name = 'c'`, &rows)
	assert.Nil(t, err, "should not error")
	assert.Equal(t, true, rows >= 0, "should gt 0")
	//Test Create
	err = dbclient.Create(nil, &Fake{
		Name:  "c",
		Value: 100,
	})

	assert.Nil(t, err, "should nil err")
	data, err := app.Execute("common.find", &fpm.BizParam{
		"table":     "fake",
		"condition": "1 = 1",
		"skip":      -1,
		"limit":     -1,
		"sort":      "id",
	})

	fmt.Printf("data: %v", data)
	assert.Nil(t, err, "should not error")

}

func TestRemoveBiz(t *testing.T) {
	app := fpm.New()

	app.Init()

	data, err := app.Execute("common.remove", &fpm.BizParam{
		"table":     "fake",
		"condition": "name = 'c'",
	})

	fmt.Printf("data: %v", data)
	assert.Nil(t, err, "should not error")

}

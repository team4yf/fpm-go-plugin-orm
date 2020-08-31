package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/team4yf/fpm-go-plugin-orm/plugins/pg"
	"github.com/team4yf/yf-fpm-server-go/fpm"
	"github.com/team4yf/yf-fpm-server-go/pkg/db"
)

//Fake 对应的实体类
type Fake struct {
	gorm.Model `json:"-"`
	Name       string `json:"name"`  // 名称
	Value      int    `json:"value"` // 状态
}

//TableName 对应表名
func (Fake) TableName() string {
	return "fake"
}
func main() {

	app := fpm.New()
	app.Init()
	dbclient, _ := app.GetDatabase("pg")
	for i := 0; i < 1000; i++ {
		go func() {
			q := db.NewQuery()
			q.AddSorter(db.Sorter{
				Sortby: "name",
				Asc:    "asc",
			}).SetTable("fake").SetCondition("name = ?", "c")
			list := make([]*Fake, 0)
			total := 0
			_ = dbclient.FindAndCount(q, &list, &total)
			app.Logger.Debugf("data: %v", list)
		}()
	}

	app.Run()

}

package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/team4yf/fpm-go-plugin-orm-pg/plugins/pg"
	"github.com/team4yf/yf-fpm-server-go/fpm"
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
	go func() {
		dbclient, _ := app.GetDatabase("pg")
		list := make([]*Fake, 0)
		total := 0
		_ = dbclient.Model(Fake{}).Condition("name = ?", "c").FindAndCount(&list, &total).Error()
		app.Logger.Debugf("data: %v", list)
	}()

	go func() {
		dbclient, _ := app.GetDatabase("pg")
		list := make([]*Fake, 0)
		total := 0
		_ = dbclient.Model(Fake{}).Condition("name = ?", "b").FindAndCount(&list, &total).Error()
		app.Logger.Debugf("data: %v", list)
	}()

	go func() {
		dbclient, _ := app.GetDatabase("pg")
		list := make([]*Fake, 0)
		total := 0
		_ = dbclient.Model(Fake{}).Condition("name = ?", "d").FindAndCount(&list, &total).Error()
		app.Logger.Debugf("data: %v", list)
	}()

	app.Run()

}

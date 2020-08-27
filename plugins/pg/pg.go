package pg

import (
	"github.com/team4yf/fpm-go-plugin-orm/plugins"
	"github.com/team4yf/yf-fpm-server-go/fpm"
	"github.com/team4yf/yf-fpm-server-go/pkg/db"

	//import the postgress
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func init() {
	fpm.Register(func(app *fpm.Fpm) {
		option := &plugins.DBSetting{}
		if err := app.FetchConfig("db", &option); err != nil {
			panic(err)
		}
		dbInstance := plugins.New(option)

		app.SetDatabase("pg", func() db.Database {
			return plugins.NewImpl(dbInstance)
		})
	})
}

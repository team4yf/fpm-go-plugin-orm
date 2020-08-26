package pg

import (
	"github.com/team4yf/yf-fpm-server-go/fpm"
	"github.com/team4yf/yf-fpm-server-go/pkg/db"

	//import the postgress
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type DBSetting struct {
	Engine   string
	User     string
	Password string
	Host     string
	Port     int
	Database string
	Charset  string
	ShowSQL  bool
}

func init() {
	fpm.Register(func(app *fpm.Fpm) {
		option := &DBSetting{}
		if err := app.FetchConfig("db", &option); err != nil {
			panic(err)
		}
		dbInstance := New(option)

		app.SetDatabase("pg", func() db.Database {
			return &pgImpl{
				db: dbInstance,
				q:  newQuery(),
			}
		})
	})
}

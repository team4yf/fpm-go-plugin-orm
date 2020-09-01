package plugins

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/team4yf/yf-fpm-server-go/pkg/db"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	locker sync.Mutex
)

//DBSetting database setting
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

//NewImpl create a new impl
func NewImpl(db *gorm.DB) db.Database {
	return &ormImpl{
		db: db,
	}
}

//ormImpl the implement of the orm
type ormImpl struct {
	locker sync.Mutex
	db     *gorm.DB
}

//New create a new instance
func New(setting *DBSetting) *gorm.DB {
	locker.Lock()
	defer locker.Unlock()
	db := CreateDb(setting)
	return db
}

//CreateDb create new instance
func CreateDb(setting *DBSetting) *gorm.DB {
	//use the config for the app
	dsn := getDbEngineDSN(setting)
	var logConf logger.Config
	if setting.ShowSQL {
		logConf = logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
			Colorful:      false,       // Disable color
		}
	} else {
		logConf = logger.Config{
			SlowThreshold: time.Second,   // Slow SQL threshold
			LogLevel:      logger.Silent, // Log level
			Colorful:      false,         // Disable color
		}
	}
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logConf,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	sqlDB.SetConnMaxLifetime(time.Minute * 30)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(50)

	return db
}

// 获取数据库引擎DSN  mysql,sqlite,postgres
func getDbEngineDSN(db *DBSetting) string {
	engine := strings.ToLower(db.Engine)
	dsn := ""
	switch engine {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&allowNativePasswords=true",
			db.User,
			db.Password,
			db.Host,
			db.Port,
			db.Database,
			db.Charset)
	case "postgres":
		dsn = fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
			db.User,
			db.Password,
			db.Host,
			db.Port,
			db.Database)
	}

	return dsn
}

//AutoMigrate migrate table from the model
func (p *ormImpl) AutoMigrate(tables ...interface{}) (err error) {
	return p.db.AutoMigrate(tables...)
}

func (p *ormImpl) Transaction(body func(db.Database) error) error {

	return p.db.Transaction(func(tx *gorm.DB) error {
		return body(&ormImpl{
			db: tx,
		})
	})
}

//OK
//Ex:
// list := make([]*Fake, 0)
// dbclient.Model(one).Sorter(db.Sorter{
// 	Sortby: "name",
// 	Asc:    "asc",
// }).Condition("name = ?", "c").Find(&list).Error()
func (p *ormImpl) Find(q *db.QueryData, result interface{}) error {
	query := p.db.Table(q.Table).Where(q.Condition, q.Arguments...)
	if len(q.Fields) > 0 {
		query = query.Select(q.Fields[0], q.Fields[1:]...)
	}
	query = query.Offset(q.Pager.Skip).Limit(q.Pager.Limit)
	if len(q.Sorter) > 0 {
		for _, sort := range q.Sorter {
			query = query.Order(sort.Sortby + " " + sort.Asc)
		}
	}
	return query.Find(result).Error
}

//OK
//Ex:
// total := 0
// err = dbclient.Model(Fake{}).Condition("name = ?", "c").Count(&total).Error()
// total is the count
func (p *ormImpl) Count(q *db.BaseData, total *int64) error {
	return p.db.Table(q.Table).Where(q.Condition, q.Arguments...).Count(total).Error
}

//OK
//Ex:
// list := make([]*Fake, 0)
// err = dbclient.Model(Fake{}).Condition("name = ?", "c").FindAndCount(&list, &total).Error()
func (p *ormImpl) FindAndCount(q *db.QueryData, result interface{}, total *int64) (err error) {
	err = p.Count(q.BaseData, total)
	if err != nil {
		return
	}
	return p.Find(q, result)
}

//OK
//Ex:
// one := &Fake{}
// err = dbclient.Model(one).Condition("name = ?", "c").First(&one).Error()
func (p *ormImpl) First(q *db.QueryData, result interface{}) error {
	query := p.db.Table(q.Table).Where(q.Condition, q.Arguments...)
	if len(q.Fields) > 0 {
		query = query.Select(q.Fields[0], q.Fields[1:]...)
	}
	query = query.Offset(q.Pager.Skip).Limit(q.Pager.Limit)
	if len(q.Sorter) > 0 {
		for _, sort := range q.Sorter {
			query = query.Order(sort.Sortby + " " + sort.Asc)
		}
	}
	return query.First(result).Error
}

//OK
//Ex:
// err = dbclient.Create(&Fake{
// 	Name:  "c",
// 	Value: 100,
// }).Error()
func (p *ormImpl) Create(_ *db.BaseData, entity interface{}) error {
	return p.db.Create(entity).Error
}

//OK:
//Ex:
// rows := 0
// err = dbclient.Model(Fake{}).Condition("name = ?", "c").Remove(&rows).Error()
func (p *ormImpl) Remove(q *db.BaseData, total *int64) (err error) {

	raw := p.db.Raw(fmt.Sprintf("UPDATE %s SET deleted_at=? WHERE %s", q.Table, q.Condition), append([]interface{}{time.Now()}, q.Arguments...)...)
	if raw.Error != nil {
		err = raw.Error
		return
	}
	err = raw.Scan(total).Error

	return
}

//OK:
//Ex:
// fields := db.CommonMap{
// 	"value": 101,
// }
// err = dbclient.Model(Fake{}).Condition("name = ?", "c").Updates(fields, &total).Error()
func (p *ormImpl) Updates(q *db.BaseData, updates db.CommonMap, rows *int64) (err error) {
	keyArr := make([]string, 0)
	valArr := make([]interface{}, 0)
	for k, v := range updates {
		if k == "" {
			continue
		}
		keyArr = append(keyArr, k+" = ?")
		valArr = append(valArr, v)
	}
	sql := fmt.Sprintf("UPDATE %s SET updated_at=?, %s WHERE deleted_at is not null and ( %s )", q.Table, strings.Join(keyArr[:], ","), q.Condition)
	params := append([]interface{}{time.Now()}, valArr...)
	params = append(params, q.Arguments...)
	raw := p.db.Raw(sql, params...)
	if raw.Error != nil {
		err = raw.Error
		return
	}
	err = raw.Scan(rows).Error

	return
}

//OK
//Ex:
//err = dbclient.Execute(`delete from fake where id = 11`, &rows).Error()
func (p *ormImpl) Execute(sql string, rows *int64) (err error) {
	d := p.db.Exec(sql)
	*rows = d.RowsAffected
	return d.Error
}

//OK:
//The result must be a struct
//Ex:
// raw := &countBody{}
// err = dbclient.Raw(`select count(1) as c from fake where id < 10`, raw).Error()
func (p *ormImpl) Raw(sql string, result interface{}) (err error) {
	raw := p.db.Raw(sql)
	if raw.Error != nil {
		err = raw.Error
		return
	}
	err = raw.Scan(result).Error

	return
}

//OK should be a struct
//Ex:
// raws := make([]*countBody, 0)
// err = dbclient.Raws(`select id as c, 1 as b from fake`, func() interface{} {
// 	return &countBody{}
// }, func(one interface{}) {
// 	raws = append(raws, one.(*countBody))
// }).Error()
func (p *ormImpl) Raws(sql string, iterator func() interface{}, appender func(interface{})) (err error) {
	d := p.db.Raw(sql)
	raws, err := d.Rows()
	if err != nil {
		return err
	}
	defer raws.Close()
	for raws.Next() {
		one := iterator()
		d.ScanRows(raws, one)
		appender(one)
	}

	return nil
}

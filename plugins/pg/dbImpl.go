package pg

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/team4yf/yf-fpm-server-go/pkg/db"
)

var (
	locker sync.Mutex
)

type queryData struct {
	condition string
	arguments []interface{}
	pager     *db.Pagination
	sorter    []db.Sorter
	err       error
	model     interface{}
}

func newQuery() *queryData {
	return &queryData{
		condition: "1=1",
		arguments: make([]interface{}, 0),
		pager: &db.Pagination{
			Skip:  0,
			Limit: -1,
		},
		sorter: make([]db.Sorter, 0),
	}
}

type pgImpl struct {
	locker sync.Mutex
	db     *gorm.DB
	q      *queryData
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
	db, err := gorm.Open(setting.Engine, dsn)
	if err != nil {
		panic(err)
	}

	db.DB().SetConnMaxLifetime(time.Minute * 5)
	db.DB().SetMaxIdleConns(20)
	db.DB().SetMaxOpenConns(500)

	db.LogMode(setting.ShowSQL)

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
func (p *pgImpl) AutoMigrate(tables ...interface{}) (err error) {
	return p.db.AutoMigrate(tables...).Error
}

func (p *pgImpl) Condition(condition string, args ...interface{}) db.Database {
	p.q.condition = condition
	p.q.arguments = args
	return p
}

func (p *pgImpl) Sorter(sorters ...db.Sorter) db.Database {
	p.q.sorter = sorters
	return p
}

func (p *pgImpl) Pager(pager *db.Pagination) db.Database {
	p.q.pager = pager
	return p
}

func (p *pgImpl) Model(model interface{}) db.Database {
	p.q = newQuery()
	p.q.model = model
	return p
}

func (p *pgImpl) Error() (err error) {
	err = p.q.err
	p.q = newQuery()
	return
}

func (p *pgImpl) Transaction(body func(db.Database) error) error {

	return p.db.Transaction(func(tx *gorm.DB) error {
		return body(&pgImpl{
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
func (p *pgImpl) Find(result interface{}) db.Database {
	//TODO sort & skip & check the result point
	query := p.db.Model(p.q.model).Where(p.q.condition, p.q.arguments)
	query = query.Offset(p.q.pager.Skip).Limit(p.q.pager.Limit)
	if len(p.q.sorter) > 0 {
		for _, sort := range p.q.sorter {
			query = query.Order(sort.Sortby + " " + sort.Asc)
		}
	}
	p.q.err = query.Find(result).Error
	return p
}

//OK
//Ex:
// total := 0
// err = dbclient.Model(Fake{}).Condition("name = ?", "c").Count(&total).Error()
// total is the count
func (p *pgImpl) Count(total *int) db.Database {
	p.q.err = p.db.Model(p.q.model).Where(p.q.condition, p.q.arguments).Count(total).Error
	return p
}

//OK
//Ex:
// list := make([]*Fake, 0)
// err = dbclient.Model(Fake{}).Condition("name = ?", "c").FindAndCount(&list, &total).Error()
func (p *pgImpl) FindAndCount(result interface{}, total *int) db.Database {
	p.Count(total)
	if p.q.err != nil {
		return p
	}
	p.Find(result)
	return p
}

//OK
//Ex:
// one := &Fake{}
// err = dbclient.Model(one).Condition("name = ?", "c").First(&one).Error()
func (p *pgImpl) First(result interface{}) db.Database {
	query := p.db.Model(p.q.model).Where(p.q.condition, p.q.arguments)
	query = query.Offset(p.q.pager.Skip).Limit(p.q.pager.Limit)
	if len(p.q.sorter) > 0 {
		for _, sort := range p.q.sorter {
			query = query.Order(sort.Sortby + " " + sort.Asc)
		}
	}
	p.q.err = query.First(result).Error
	return p
}

//OK
//Ex:
// err = dbclient.Create(&Fake{
// 	Name:  "c",
// 	Value: 100,
// }).Error()
func (p *pgImpl) Create(entity interface{}) db.Database {
	if p.q == nil {
		p.q = newQuery()
		p.q.model = entity
	}
	p.q.err = p.db.Create(entity).Error
	return p
}

//OK
//Ex:
// rows := 0
// err = dbclient.Model(Fake{}).Condition("name = ?", "c").Remove(&rows).Error()
func (p *pgImpl) Remove(total *int) db.Database {
	d := p.db.Where(p.q.condition, p.q.arguments).Delete(p.q.model)
	*total = (int)(d.RowsAffected)
	p.q.err = d.Error
	return p
}

//OK:
//Ex:
// fields := db.CommonMap{
// 	"value": 101,
// }
// err = dbclient.Model(Fake{}).Condition("name = ?", "c").Updates(fields, &total).Error()
func (p *pgImpl) Updates(updates db.CommonMap, rows *int) db.Database {
	q := p.db.Model(p.q.model).Where(p.q.condition, p.q.arguments).Updates(updates)
	p.q.err = q.Error
	*rows = (int)(q.RowsAffected)
	return p
}

//OK
//Ex:
//err = dbclient.Execute(`delete from fake where id = 11`, &rows).Error()
func (p *pgImpl) Execute(sql string, rows *int) db.Database {
	d := p.db.Exec(sql)
	*rows = (int)(d.RowsAffected)
	p.q = newQuery()
	p.q.err = d.Error
	return p
}

//OK:
//The result must be a struct
//Ex:
// raw := &countBody{}
// err = dbclient.Raw(`select count(1) as c from fake where id < 10`, raw).Error()
func (p *pgImpl) Raw(sql string, result interface{}) db.Database {
	raw := p.db.Raw(sql)
	if raw.Error != nil {
		p.q.err = raw.Error
		return p
	}
	p.q.err = raw.Scan(result).Error

	return p
}

//OK should be a struct
//Ex:
// raws := make([]*countBody, 0)
// err = dbclient.Raws(`select id as c, 1 as b from fake`, func() interface{} {
// 	return &countBody{}
// }, func(one interface{}) {
// 	raws = append(raws, one.(*countBody))
// }).Error()
func (p *pgImpl) Raws(sql string, iterator func() interface{}, appender func(interface{})) db.Database {
	d := p.db.Raw(sql)
	raws, err := d.Rows()
	if err != nil {
		p.q.err = err
		return p
	}
	defer raws.Close()
	for raws.Next() {
		one := iterator()
		d.ScanRows(raws, one)
		appender(one)
	}

	return p
}

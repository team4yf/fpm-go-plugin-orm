package plugins

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
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

type MigrationHistory struct {
	gorm.Model  `json:"-"`
	Version     string
	Description string
	Script      string
	InstalledAt time.Time
	Success     int
}

func (h *MigrationHistory) TableName() string {
	return "migration_histories"
}

//NewImpl create a new impl
func NewImpl(db *gorm.DB) db.Database {
	return &ormImpl{
		db: db,
	}
}

//ormImpl the implement of the orm
type ormImpl struct {
	// locker sync.Mutex
	db *gorm.DB
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
	sqlDB, _ := db.DB()
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
	migrationHistory := MigrationHistory{}
	if err = p.db.AutoMigrate(&migrationHistory); err != nil {
		return
	}
	if err = p.db.AutoMigrate(tables...); err != nil {
		return
	}
	// get migration scripts from folder

	query := db.NewQuery()
	query.SetTable(migrationHistory.TableName())
	var count int64
	if err = p.Count(query.BaseData, &count); err != nil {
		return
	}
	if count > 0 {
		query.AddSorter(db.Sorter{
			Sortby: "installed_at",
			Asc:    "desc",
		})
		err = p.First(query, &migrationHistory)
	}
	reScript, _ := regexp.Compile(`^V(\d|\.)+`)
	// fetch scripts
	scripts := []string{}
	if files, ex := ioutil.ReadDir("migrations"); ex != nil {
		return ex
	} else {
		for _, f := range files {
			if count == 0 || reScript.FindString(f.Name()) > reScript.FindString(migrationHistory.Script) {
				scripts = append(scripts, f.Name())
			}
		}
	}
	if len(scripts) == 0 {
		return
	}
	reVersion, _ := regexp.Compile(`^V\d+`)
	reDesc, _ := regexp.Compile(`__(\w|\s|_|-|\+)+`)
	sort.Strings(scripts)
	for _, s := range scripts {
		var raw []byte
		if raw, err = os.ReadFile(fmt.Sprintf("migrations%s%s", string(filepath.Separator), s)); err != nil {
			return
		}
		if err = p.db.Transaction(func(tx *gorm.DB) (ex error) {
			// run script
			if ex = tx.Exec(string(raw)).Error; ex != nil {
				return
			}
			// add record
			if ex = tx.Create(&MigrationHistory{
				Version:     reVersion.FindString(s),
				Description: reDesc.FindString(s),
				Script:      s,
				InstalledAt: time.Now(),
				Success:     1,
			}).Error; ex != nil {
				return
			}
			return
		}); err != nil {
			return
		}
	}
	return
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
func (p *ormImpl) Find(q *db.QueryData, result interface{}) (err error) {

	switch result.(type) {
	case *[]map[string]interface{}:
		data, e := p.FindObject(q)
		list := result.(*[]map[string]interface{})
		*list = append(*list, data...)
		if e != nil {
			return e
		}
		return
	}
	query := p.db.Table(q.Table).Where(fmt.Sprintf("(%s) and deleted_at is null", q.Condition), q.Arguments...)
	if len(q.Fields) > 0 {
		fields := make([]interface{}, len(q.Fields))
		for i, v := range q.Fields {
			fields[i] = v
		}
		query = query.Select(fields[0], fields[1:]...)
	}
	query = query.Offset(q.Pager.Skip).Limit(q.Pager.Limit)
	if len(q.Sorter) > 0 {
		for _, sort := range q.Sorter {
			query = query.Order(sort.Sortby + " " + sort.Asc)
		}
	}
	return query.Find(result).Error

}

func (p *ormImpl) FindObject(q *db.QueryData) (data []map[string]interface{}, err error) {
	query := p.db.Table(q.Table).Where(fmt.Sprintf("(%s) and deleted_at is null", q.Condition), q.Arguments...)
	if len(q.Fields) > 0 {
		fields := make([]interface{}, len(q.Fields))
		for i, v := range q.Fields {
			fields[i] = v
		}
		query = query.Select(fields[0], fields[1:]...)
	}
	query = query.Offset(q.Pager.Skip).Limit(q.Pager.Limit)
	if len(q.Sorter) > 0 {
		for _, sort := range q.Sorter {
			query = query.Order(sort.Sortby + " " + sort.Asc)
		}
	}
	rows, err := query.Rows()
	if err != nil {
		return
	}
	defer rows.Close()
	data = make([]map[string]interface{}, 0)
	cols, _ := rows.Columns()

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err = rows.Scan(columnPointers...); err != nil {
			return
		}
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}
		data = append(data, m)

	}
	return
}

//OK
//Ex:
// total := 0
// err = dbclient.Model(Fake{}).Condition("name = ?", "c").Count(&total).Error()
// total is the count
func (p *ormImpl) Count(q *db.BaseData, total *int64) error {
	return p.db.Table(q.Table).Where(fmt.Sprintf("(%s) and deleted_at is null", q.Condition), q.Arguments...).Count(total).Error
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
func (p *ormImpl) First(q *db.QueryData, result interface{}) (err error) {
	query := p.db.Table(q.Table)
	if len(q.Fields) > 0 {
		fields := make([]interface{}, len(q.Fields))
		for i, v := range q.Fields {
			fields[i] = v
		}
		query = query.Select(fields[0], fields[1:]...)
	}
	query = query.Offset(0).Limit(1)
	if len(q.Sorter) > 0 {
		for _, sort := range q.Sorter {
			query = query.Order(sort.Sortby + " " + sort.Asc)
		}
	}
	query.Where(fmt.Sprintf("(%s) and deleted_at is null", q.Condition), q.Arguments...)
	switch result.(type) {
	case *map[string]interface{}:
		rows, e := query.Rows()
		if e != nil {
			return e
		}
		defer rows.Close()

		cols, _ := rows.Columns()

		if rows.Next() {
			columns := make([]interface{}, len(cols))
			columnPointers := make([]interface{}, len(cols))
			for i := range columns {
				columnPointers[i] = &columns[i]
			}
			if err = rows.Scan(columnPointers...); err != nil {
				return
			}
			m := make(map[string]interface{})
			for i, colName := range cols {
				val := columnPointers[i].(*interface{})
				m[colName] = *val
			}
			p := result.(*map[string]interface{})
			*p = m
		}
		return nil

	}
	return query.First(result).Error
}

//OK
//Ex:
// err = dbclient.Create(&Fake{
// 	Name:  "c",
// 	Value: 100,
// }).Error()
func (p *ormImpl) Create(q *db.BaseData, entity interface{}) error {
	d := p.db.Table(q.Table)
	//判断传入的entity的类型，如果是结构体或者结构体指针，则直接创建
	objType := reflect.TypeOf(entity)
	if objType.Kind() == reflect.Ptr {
		//指针
		objType = objType.Elem()
	}
	if objType.Kind() == reflect.Struct {
		return d.Create(entity).Error
	}

	var e map[string]interface{}
	switch entity.(type) {
	case *map[string]interface{}:
		one := entity.(*map[string]interface{})
		e = *one
		return nil
	case map[string]interface{}:
		e = entity.(map[string]interface{})
	case interface{}:
		//通过json转义过来的空接口类型，本身可能是 map 类型
		e = entity.(map[string]interface{})
	default:
		return errors.New("unknown data type")
	}
	//TODO: do sql
	sql := `INSERT INTO "%s" ("created_at","updated_at","deleted_at",%s) 
	VALUES (?,?,NULL,%s) RETURNING "id"`
	keys := make([]string, 0)
	vals := make([]string, 0)
	for k, v := range e {
		if k == "updateAt" || k == "createAt" || k == "createat" || k == "updateat" {
			continue
		}
		keys = append(keys, "\""+k+"\"")
		switch v.(type) {
		case string:
			vals = append(vals, "'"+v.(string)+"'")
		case float64:
			f := v.(float64)
			if int64(f*1000)%1000 == 0 {
				// it's a int
				vals = append(vals, fmt.Sprintf("%d", int64(f)))
			} else {
				vals = append(vals, fmt.Sprintf("%f", f))
			}
		default:
			vals = append(vals, "\""+fmt.Sprintf("%v", v)+"\"")
		}

	}
	sql = fmt.Sprintf(sql, q.Table, strings.Join(keys, ","), strings.Join(vals, ","))

	now := time.Now()

	if err := p.db.Exec(sql, now, now).Error; err != nil {
		return err
	}

	return nil
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
	if *total == 0 {
		*total = 1
	}
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
		if k == "updateAt" {
			continue
		}
		keyArr = append(keyArr, k+" = ?")
		switch v.(type) {
		case float64:
			f := v.(float64)
			if int64(f*1000)%1000 == 0 {
				// it's a int
				valArr = append(valArr, int64(f))
				continue
			}
		}
		valArr = append(valArr, v)
	}
	sql := fmt.Sprintf("UPDATE %s SET updated_at=?, %s WHERE deleted_at is null and ( %s )", q.Table, strings.Join(keyArr[:], ","), q.Condition)
	params := append([]interface{}{time.Now()}, valArr...)
	params = append(params, q.Arguments...)
	raw := p.db.Raw(sql, params...)
	if raw.Error != nil {
		err = raw.Error
		return
	}
	//TODO: 这里丢失了行数
	err = raw.Scan(rows).Error
	if *rows == 0 {
		*rows = 1
	}
	return
}

//OK
//Ex:
//err = dbclient.Execute(`delete from fake where id = 11`, &rows).Error()
func (p *ormImpl) Execute(sql string, rows *int64) (err error) {
	d := p.db.Exec(sql)
	err = d.Error
	if err != nil {
		return
	}
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

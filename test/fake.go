package fake

import (
	"github.com/jinzhu/gorm"
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

package amber

import (
  "github.com/jinzhu/gorm"
)

type Model struct {

}

func (m *Model) DB() *gorm.DB {
	return db
}
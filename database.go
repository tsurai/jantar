package amber

import (
  "github.com/jinzhu/gorm"
  _ "github.com/lib/pq"
)

type DatabaseConfig struct {
  Driver    string
  Username  string
  Database  string
}

type database struct {
  gorm.DB
  initialized bool
  config      *DatabaseConfig
}

var DB database

func (db *database) initialize(config *DatabaseConfig) (err error) {
  gdb, err := gorm.Open(config.Driver, "user=" + config.Username + " dbname=" + config.Database + " sslmode=disable")
  if err != nil {
    return err
  }

  db.DB = gdb
  db.config = config
  db.initialized = true

  return err
}
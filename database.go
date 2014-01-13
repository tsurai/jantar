package amber

import (
  "fmt"
  "github.com/jinzhu/gorm"
  //_ "github.com/lib/pq"
  _ "github.com/mattn/go-sqlite3"
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
  var gdb gorm.DB

  // just a dummy for two testing environments
  // final switch will be a lot more dynamic
  switch config.Driver {
  case "postgres":
    gdb, err = gorm.Open(config.Driver, "user=" + config.Username + " dbname=" + config.Database + " sslmode=disable")
  case "sqlite3":
    gdb, err = gorm.Open(config.Driver, config.Database + ".db")
  default:
    err = fmt.Errorf("Unknown database driver")
  }

  if err != nil {
    return err
  }

  db.DB = gdb
  db.config = config
  db.initialized = true

  return err
}
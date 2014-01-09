package amber

import (
  "fmt"
  "github.com/jinzhu/gorm"
  _ "github.com/lib/pq"
)

var db *gorm.DB

func (a *Amber) InitDatabase(user string, dbname string) *gorm.DB {
  var err error
  
  database, err := gorm.Open("postgres", "user=" + user + " dbname=" + dbname + " sslmode=disable")
  if err != nil {
    panic(fmt.Sprintf("Got error when connect database, the error is '%v'", err))
  }

  db = &database
  return db
}
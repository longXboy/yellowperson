package db

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"yellowPerson/util/config"
)

var DB *xorm.Engine

func InitDB() error {
	dbuser, err := config.GetConfigString("db.user")
	if err != nil {
		return err
	}
	dbpasswd, err := config.GetConfigString("db.psw")
	if err != nil {
		return err
	}
	dbaddr, err := config.GetConfigString("db.addr")
	if err != nil {
		return err
	}
	dbname, err := config.GetConfigString("db.name")
	if err != nil {
		return err
	}
	connstr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4", dbuser, dbpasswd, dbaddr, dbname)
	DB, err = xorm.NewEngine("mysql", connstr)
	if err != nil {
		return err
	}
	err = DB.Ping()
	if err != nil {
		return err
	}
	DB.SetMaxConns(50)
	DB.SetMaxIdleConns(5)
	DB.SetMaxOpenConns(50)
	DB.SetMapper(core.SameMapper{})
	DB.ShowErr = true
	return nil
}

func FiniDB() {
	DB.Close()
}

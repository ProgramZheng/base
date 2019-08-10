package model

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/spf13/viper"
)

var (
	db  *gorm.DB
	err error
)

type Result struct {
	Value interface{}
	Error error
}

func init() {
	//?parseTime=true for the database table column type is TIMESTAMP
	setting := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=true", viper.Get("DB_USERNAME"), viper.Get("DB_PASSWORD"), viper.Get("DB_HOST"), viper.Get("DB_PORT"), viper.Get("DB_DATABASE"))

	db, err = gorm.Open("mysql", setting)

	// defer db.Close()
	if err != nil {
		log.Println("DataBase error:", err)
	}
	// err = db.Ping()
	// if err != nil {
	// 	log.Println("DataBase Ping error:", err)
	// }
}

//Add is add model to database
// interface can't get origin variable only get variable at memory location
func Add(model interface{}) (result Result) {
	//create table for the struct
	result = Result{
		Value: nil,
		Error: nil,
	}
	db.AutoMigrate(model)
	if dbc := db.Create(model); dbc.Error != nil {
		//error
		result.Error = dbc.Error
	}
	return
}

//GetForID is get first model For ID from database
func GetForID(model interface{}, id int) (result interface{}) {
	result = db.First(model, id)
	return
}

//Get is get first model to database
func Get(model interface{}, where interface{}) (result interface{}) {
	result = db.Where(where).Find(model)
	return
}

//Save is get first model then update data to database
func Save(model interface{}, id int, update interface{}) (result interface{}) {
	result = db.First(model, id).Updates(update)
	return
}

//Del is del find model to database
func Del(model interface{}, id int) (result interface{}) {
	result = db.Delete(model, id)
	return
}

package mysql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("mysql", "user:1234@tcp(localhost:3306)/breakfast?charset=utf8")
	if err != nil {
		panic(err)
	}

	fmt.Println("sql.DB 結構已建立")

	err = DB.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("資料庫連線成功")
}

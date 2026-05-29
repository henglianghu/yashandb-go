package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/yashan-technologies/yashandb-go"
	example "github.com/yashan-technologies/yashandb-go/_examples"
)

func main() {
	dsn := example.GetDsn()
	db, err := sql.Open("yasdb", dsn)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	_, err = db.Exec("drop table if exists emoji_example")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("create table emoji_example(c1 char(20),c2 varchar(20))")
	if err != nil {
		fmt.Println(err)
		return
	}

	c1 := "😇"
	c2 := "😮"
	_, err = db.Exec("insert into emoji_example(c1,c2) values(?,?)", c1, c2)
	if err != nil {
		fmt.Println(err)
		return
	}
	rows, err := db.Query("select c1,c2 from emoji_example")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var d1 string
		var d2 string
		rows.Scan(&d1, &d2)
		d1 = strings.TrimSpace(d1)
		if c1 != d1 || c2 != d2 {
			fmt.Println("emoji doesn't work correctly")
			os.Exit(1)
		}
		fmt.Println(d1, d2)
	}
}

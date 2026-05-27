package main

import (
	"database/sql"
	"fmt"
	"os"

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

	_, err = db.Exec("drop table if exists raw_example;")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("create table raw_example(c1 raw(17))  ;")
	if err != nil {
		fmt.Println(err)
		return
	}

	c1 := []byte("你好Yashandb！")
	result, err := db.Exec("insert into raw_example(c1) values(?);  ", c1)
	if err != nil {
		fmt.Println(err)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("affectRows:", rowsAffected)

	rows, err := db.Query("select c1 from raw_example  ;  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	names, err := rows.Columns()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("columnNames:", names)
	defer rows.Close()
	for rows.Next() {
		var oc1 []byte
		rows.Scan(&oc1)
		if string(oc1) != string(c1) {
			fmt.Println("raw doesn't work correctly")
			os.Exit(1)
		}
		fmt.Println(string(oc1))
	}
}

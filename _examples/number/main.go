package main

import (
	"database/sql"
	"fmt"

	_ "github.com/yashan-technologies/yashandb-go"
)

func main() {
	dsn := "sys/Cod-2022@127.0.0.1:1688?number_as_string=true"

	db, err := sql.Open("yasdb", dsn)
	if err != nil {
		fmt.Println("failed to connect yashandb, err:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec("drop table if exists number_example")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("create table number_example(n NUMBER(20,0))")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("insert into number_example values(1752174832254285824)")
	if err != nil {
		fmt.Println(err)
		return
	}

	rows, err := db.Query("select * from number_example")
	if err != nil {
		fmt.Println(err)
		return
	}
	for rows.Next() {
		var value string
		err = rows.Scan(&value)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%v\n", value)
	}
}

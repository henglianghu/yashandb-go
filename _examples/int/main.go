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

	_, err = db.Exec("drop table if exists int_example;")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("create table int_example(id int)  ;")
	if err != nil {
		fmt.Println(err)
		return
	}

	id := 10
	result, err := db.Exec("insert into int_example(id) values(?);  ", id)
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

	rows, err := db.Query("select id from int_example  ;  ")
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
		var idInDB int
		rows.Scan(&idInDB)
		if idInDB != id {
			fmt.Println("int doesn't work correctly")
			os.Exit(1)
		}
		fmt.Println(idInDB)
	}
}

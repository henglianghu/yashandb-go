package main

import (
	"database/sql"
	"fmt"

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

	_, err = db.Exec("drop table if exists null_example;")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("create table null_example(id int,c1 int,c2 boolean,c3 date,c4 float) ;")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("insert into null_example(id) values(?);  ", 10)
	if err != nil {
		fmt.Println(err)
		return
	}

	rows, err := db.Query("select c1, c2, c3, c4 from null_example;")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			c1 sql.NullInt32
			c2 sql.NullBool
			c3 sql.NullTime
			c4 sql.NullFloat64
		)
		if err := rows.Scan(&c1, &c2, &c3, &c4); err != nil {
			fmt.Println(err)
			return
		}
		if c1.Valid != false || c2.Valid != false || c3.Valid != false || c4.Valid != false {
			fmt.Println("null doesn't work correctly")
			return
		}
	}
}

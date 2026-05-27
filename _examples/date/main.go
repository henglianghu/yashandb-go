package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

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

	_, err = db.Exec("drop table if exists date_example")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("create table date_example(date1 DATE, date2 TIMESTAMP)")
	if err != nil {
		fmt.Println(err)
		return
	}

	date1 := time.Now()
	date2 := date1.AddDate(-1, 0, 1)
	_, err = db.Exec("insert into date_example(date1,date2) values(:date1,:date2)", date1, date2)
	if err != nil {
		fmt.Println(err)
		return
	}

	rows, err := db.Query("select date1,date2 from date_example")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var d1 time.Time
		var d2 time.Time
		rows.Scan(&d1, &d2)
		if date1.Unix() != d1.Unix() || date2.UnixMicro() != d2.UnixMicro() {
			fmt.Println(date1, date2)
			fmt.Println(d1, d2)
			fmt.Println("date doesn't work correctly")
			os.Exit(1)
		}
		fmt.Println(d1, d2)
	}
}

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

	_, err = db.Exec("drop table if exists xml_example")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("create table xml_example(c1 int,c2 xmltype)")
	if err != nil {
		fmt.Println(err)
		return
	}

	c1 := 1
	c2 := `<employee><id>2</id><name>hahaha</name></employee>`
	_, err = db.Exec("insert into xml_example(c1,c2) values(?,?)", c1, c2)
	if err != nil {
		fmt.Println(err)
		return
	}
	rows, err := db.Query("select c1,c2 from xml_example")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var d1 int
		var d2 string
		if err := rows.Scan(&d1, &d2); err != nil {
			fmt.Println(err)
			return
		}
		if c1 != d1 || c2 != d2 {
			fmt.Println("xml doesn't work correctly")
			os.Exit(1)
		}
		fmt.Println(d1, d2)
	}
}

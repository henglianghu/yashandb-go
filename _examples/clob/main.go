package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

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

	_, err = db.Exec("drop table if exists clob_example")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("create table clob_example(clob1 clob, clob2 clob)")
	if err != nil {
		fmt.Println(err)
		return
	}

	clob1 := "你好,YashanDB！"
	clob2 := getclob()
	_, err = db.Exec("insert into clob_example(clob1,clob2) values(?,?)", clob1, clob2)
	if err != nil {
		fmt.Println(err)
		return
	}

	rows, err := db.Query("select clob1,clob2 from clob_example")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var c1 string
		var c2 string
		rows.Scan(&c1, &c2)
		if clob1 != c1 || clob2 != c2 {
			fmt.Println("clob doesn't work correctly")
			os.Exit(1)
		}
		fmt.Println(c1, c2)
	}
}

// Over 8192bytes
func getclob() string {
	str := ""
	for i := 0; i < 9000; i++ {
		str += strconv.Itoa(i) + ","
		if len(str) > 9000 {
			break
		}
	}
	return str
}

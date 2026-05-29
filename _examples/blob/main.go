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

	_, err = db.Exec("drop table if exists blob_example")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("create table blob_example(blob1 blob, blob2 blob)")
	if err != nil {
		fmt.Println(err)
		return
	}

	blob1 := []byte("你好YahsanDB！")
	blob2 := getBlobBytes()
	_, err = db.Exec("insert into blob_example(blob1,blob2) values(?,?)", blob1, blob2)
	if err != nil {
		fmt.Println(err)
		return
	}

	rows, err := db.Query("select blob1,blob2 from blob_example")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var b1 []byte
		var b2 []byte
		rows.Scan(&b1, &b2)
		if string(blob1) != string(b1) || string(blob2) != string(b2) {
			fmt.Println("BLOB doesn't work correctly")
			os.Exit(1)
		}
		fmt.Println(string(b1), string(b2))
	}
}

// Over 8192bytes
func getBlobBytes() []byte {
	str := ""
	for i := 0; i < 9000; i++ {
		str += strconv.Itoa(i) + ","
		if len(str) > 9000 {
			break
		}
	}
	return []byte(str)
}

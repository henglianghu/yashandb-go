package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

	_, err = db.Exec("drop table if exists context_example")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("create table context_example(blob1 blob, blob2 blob)")
	if err != nil {
		fmt.Println(err)
		return
	}

	blob1 := []byte("你好YahsanDB！")
	blob2 := []byte(strings.Repeat(string(blob1), 1000000))
	ctx, cancel := context.WithTimeout(context.Background(), 10000000*time.Millisecond)
	_, err = db.ExecContext(ctx, "insert into context_example(blob1,blob2) values(?,?)", blob1, blob2)
	if err != nil {
		fmt.Println(err)
		return
	}
	cancel()
}

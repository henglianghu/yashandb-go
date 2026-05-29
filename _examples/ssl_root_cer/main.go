package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/yashan-technologies/yashandb-go"
)

func main() {
	dsn := "sys/Cod-2022@127.0.0.1:1688?ssl_root_cer=/opt/ycm/ycm/etc/db_crt_tmp/minidb_ssl/root.crt"
	db, err := sql.Open("yasdb", dsn)
	if err != nil {
		fmt.Println("failed to connect yashandb, err:", err)
		return
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		fmt.Println("ping failed, err: ", err)
	}
	rows, err := db.Query("select sid from v$session limit 1")
	if err != nil {
		fmt.Println(err)
		return
	}
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(name)
	}
}

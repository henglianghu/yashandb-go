package main

import (
	"database/sql"
	"fmt"

	_ "github.com/yashan-technologies/yashandb-go"
	example "github.com/yashan-technologies/yashandb-go/_examples"
)

func getYasdbConn(dsn string) (*sql.DB, error) {
	return sql.Open("yasdb", dsn)
}

func main() {
	dsn := example.GetDsn()
	// dsn = "sys/Cod-2022@127.0.0.1:1688"
	dsn = "/home/yangdeliu/gitlab/anchorbase/install/data/yasdb/mn-1-1"
	db, err := sql.Open("yasdb", dsn)
	if err != nil {
		fmt.Println("failed to connect yashandb, err:", err)
		return
	}
	defer db.Close()
	rows, err := db.Query("select name, value from v$parameter where name='RUN_LOG_FILE_PATH'")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("select name, value from v$parameter where name='RUN_LOG_FILE_PATH'")
	for rows.Next() {
		var name, value string
		err = rows.Scan(&name, &value)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(name, "=>", value)
	}
}

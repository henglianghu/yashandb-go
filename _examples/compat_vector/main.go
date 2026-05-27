package main

import (
	"database/sql"
	"fmt"

	_ "github.com/yashan-technologies/yashandb-go"
)

func main() {
	dsn := "sys/Cod-2022@127.0.0.1:1688?compat_vector=mysql"
	db, err := sql.Open("yasdb", dsn)
	if err != nil {
		fmt.Println("failed to connect yashandb, err:", err)
		return
	}
	defer db.Close()
	rows, err := db.Query("select name, value from v$parameter where name='COMPAT_VECTOR'")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("select name, value from v$parameter where name='COMPAT_VECTOR'")
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

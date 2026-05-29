package main

import (
	"context"
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
	db, err := sql.Open("yasdb", dsn)
	if err != nil {
		fmt.Println("failed to connect yashandb, err:", err)
		return
	}
	defer db.Close()

	if err := bindParameterByName(db); err != nil {
		fmt.Println(err)
		return
	}

	if err := bindParameterByPosition(db); err != nil {
		fmt.Println(err)
		return
	}
}

func bindParameterByName(db *sql.DB) error {
	rows, err := db.QueryContext(context.Background(), "select name, value from v$parameter where name=:c1", sql.NamedArg{Name: "c1", Value: "RUN_LOG_FILE_PATH"})
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var name, value string
		err = rows.Scan(&name, &value)
		if err != nil {
			return err
		}
		fmt.Println("bindParameterByName", name, "=>", value)
	}
	return nil
}

func bindParameterByPosition(db *sql.DB) error {
	rows, err := db.QueryContext(context.Background(), "select name, value from v$parameter where name=?", "RUN_LOG_FILE_PATH")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var name, value string
		err = rows.Scan(&name, &value)
		if err != nil {
			return err
		}
		fmt.Println("bindParameterByPosition", name, "=>", value)
	}
	return nil
}

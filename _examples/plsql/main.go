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
	db, err := sql.Open("yasdb", dsn)
	if err != nil {
		fmt.Println("failed to connect yashandb, err:", err)
		return
	}
	defer db.Close()
	_, err = db.Exec(`CREATE  OR    REPLACE FUNCTION ya_proc1(i INT) RETURN VARCHAR
    IS
    BEGIN
    CASE i
    WHEN 1 THEN
    RETURN 'hello';
    WHEN 2 THEN
    RETURN 'world';
    END CASE;
    END ya_proc1;`)
	if err != nil {
		fmt.Println(err)
		return
	}

	rows, err := db.Query("SELECT ya_proc1(1)||' '||ya_proc1(2) FROM dual")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var str string
		err = rows.Scan(&str)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("ya_proc1:", str)
	}

	_, err = db.Exec(`CREATE OR REPLACE PROCEDURE ya_proc2(i INT) IS
    BEGIN
    CASE i
    WHEN 1 THEN
    DBMS_OUTPUT.PUT_LINE('hello');
    WHEN 2 THEN
    DBMS_OUTPUT.PUT_LINE('world');
    END CASE;
    END ya_proc2;`)
	if err != nil {
		fmt.Println(err)
		return
	}
}

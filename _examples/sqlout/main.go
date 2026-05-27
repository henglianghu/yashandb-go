package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	yasdbgo "github.com/yashan-technologies/yashandb-go"
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

	_, err = db.Exec("drop table if exists out_example")
	if err != nil {
		fmt.Println(err)
		return
	}
	db.Exec("drop sequence out_sequence")
	_, err = db.Exec("drop trigger if exists out_trigger")
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = db.Exec("create table out_example(id int primary key, c2 varchar(20),c3 clob,c4 blob,c5 date,c6 float,c7 double,c8 number)")
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = db.Exec("CREATE SEQUENCE out_sequence increment by 1 start with 1 nomaxvalue nocycle nocache")
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = db.Exec(`
    create trigger out_trigger before insert on out_example for each row when (new.id is null)
    begin 
    select out_sequence.nextval into:new.id from dual;
    end;
    `)
	if err != nil {
		fmt.Println(err)
		return
	}

	var (
		id    int
		outC2 string
		outC3 string
		outC4 []byte
		outC5 time.Time
		outC6 float32
		outC7 float64
		outC8 float64
	)
	c2 := "varchar(20)"
	c3 := strings.Repeat("你好，YashanDB！", 10) + "....."
	c4 := []byte(c3)
	c5 := time.Now()
	c6 := float32(6)
	c7 := float64(7)
	c8 := float64(8.5)

	outC2BindValue, err := yasdbgo.NewOutputBindValue(&outC2, yasdbgo.WithTypeVarchar(), yasdbgo.WithBindSize(12))
	if err != nil {
		fmt.Println(err)
		return
	}
	outC3BindValue, err := yasdbgo.NewOutputBindValue(&outC3, yasdbgo.WithTypeClob())
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec(
		"insert into out_example(c2,c3,c4,c5,c6,c7,c8) values(?,?,?,?,?,?,?) returning id,c2,c3,c4,c5,c6,c7,c8 into ?,?,?,?,?,?,?,?",
		c2,
		c3,
		c4,
		c5,
		c6,
		c7,
		c8,
		sql.Out{Dest: &id},
		sql.Out{Dest: outC2BindValue},
		sql.Out{Dest: outC3BindValue},
		sql.Out{Dest: &outC4},
		sql.Out{Dest: &outC5},
		sql.Out{Dest: &outC6},
		sql.Out{Dest: &outC7},
		sql.Out{Dest: &outC8},
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	if id != 1 || outC2 != c2 || outC3 != c3 || string(outC4) != string(c4) || outC5.Unix() != c5.Unix() || outC6 != c6 || outC7 != c7 || outC8 != c8 {
		fmt.Println("sql output doesn't work correctly")
		os.Exit(1)
	}

	rows, err := db.Query("select id,c2,c3,c4,c5,c6,c7,c8 from out_example")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			d1 int
			d2 string
			d3 string
			d4 []byte
			d5 time.Time
			d6 float32
			d7 float64
			d8 float64
		)
		rows.Scan(&d1, &d2, &d3, &d4, &d5, &d6, &d7, &d8)
		if 1 != d1 || c2 != d2 || c3 != d3 || string(c4) != string(d4) || c5.Unix() != d5.Unix() || c6 != d6 || c7 != d7 || c8 != d8 {
			fmt.Println("sql output doesn't work correctly")
			os.Exit(1)
		}
		fmt.Println(d1, d2, d3, d5, d6, d7, d8)
	}
}

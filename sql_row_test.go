package yasdb

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestColumnTypePrecisionScale(t *testing.T) {
	db, err := sql.Open("yasdb", testDsn)
	if err != nil {
		t.Fatalf("%s%v", NormalConnErr, err)
	}
	defer db.Close()

	testCases := []struct {
		caseName  string
		precision int64
		scale     int64
		ok        bool
	}{
		{"case1", 3, -2, true},
		{"case2", 6, -3, true},
		{"case3", 5, 0, true},
		{"case4", 10, -1, true},
	}
	for _, v := range testCases {
		t.Run(v.caseName, func(t *testing.T) {
			queryStr := fmt.Sprintf("SELECT CAST(31401.465646 AS NUMBER(%d,%d)) FROM DUAL;", v.precision, v.scale)
			rows, err := db.Query(queryStr)
			if err != nil {
				t.Fatal(err)
			}
			rowTypes, err := rows.ColumnTypes()
			if err != nil {
				t.Fatal(err)
			}
			for _, info := range rowTypes {
				precision, scale, ok := info.DecimalSize()
				if ok != v.ok || precision != v.precision || scale != v.scale {
					t.Fatalf("TestColumnTypePrecisionScale %q failed；query result:%v, precision:%v, scale:%v; want result:%v, precision:%v, scale:%v ", queryStr, ok, precision, scale, true, v.precision, v.scale)
				}
			}
		})
	}
}

func TestColumnTypeNullable(t *testing.T) {
	db, err := sql.Open("yasdb", testDsn)
	if err != nil {
		t.Fatalf("%s%v", NormalConnErr, err)
	}
	defer db.Close()

	testCases := []struct {
		caseName string
		rowType  string
		nullable bool
	}{
		{"case1", "char20()", true},
		{"case2", "varchar(20)", false},
		{"case3", "int", true},
		{"case4", "float", false},
	}
	for _, v := range testCases {
		t.Run(v.caseName, func(t *testing.T) {
			tableName := fmt.Sprintf("TestColumnTypeNullable_%s", v.caseName)
			if _, err := db.Exec(fmt.Sprintf("drop table if exists %s", tableName)); err != nil {
				t.Fatal(err)
			}
			createSql := fmt.Sprintf("create table %s (name char(20))", tableName)
			if !v.nullable {
				createSql = fmt.Sprintf("create table %s (name char(20) not null)", tableName)
			}
			_, err := db.Exec(createSql)
			if err != nil {
				t.Fatal(err)
			}

			rows, err := db.Query(fmt.Sprintf("select * from %s", tableName))
			if err != nil {
				t.Fatal(err)
			}
			rowTypes, err := rows.ColumnTypes()
			if err != nil {
				t.Fatal(err)
			}
			for _, info := range rowTypes {
				nullbale, _ := info.Nullable()
				if nullbale != v.nullable {
					t.Fatalf("TestColumnTypeNullable %q failed；query nullable result:%v ; want result:%v", createSql, nullbale, v.nullable)
				}
			}
		})
	}
}

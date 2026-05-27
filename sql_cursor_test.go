package yasdb

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
)

func TestCursorFetch(t *testing.T) {
	db, err := sql.Open("yasdb", testDsn)
	if err != nil {
		t.Fatalf("error connecting: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}

	sqls := []string{
		"drop table if exists t2_cursor",
		"create table t2_cursor (col1 int,col2 varchar(10));",
		"insert into t2_cursor values (1,'9999');",
		"insert into t2_cursor values (10,'99999');",
		`
		create or replace function selectcursor(c1 in int) 
		return sys_refcursor as 
		c2 sys_refcursor; 
		begin 
			open c2 for select * from t2_cursor where col1 = c1; 
			return c2;
		end;
		`,
	}
	for _, s := range sqls {
		execute(t, conn, s)
	}

	cursorQuery := "select selectcursor(1) C from dual"
	rows, err := conn.QueryContext(ctx, cursorQuery)
	if err != nil {
		t.Fatal(err)
	}

	// 校验cursor列元数据
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		t.Fatal(err)
	}
	expectedColTypes := [][]string{
		{"C", "CURSOR"},
	}
	actualColTypes := [][]string{}
	for _, col := range colTypes {
		actualColTypes = append(actualColTypes, []string{
			col.Name(),
			col.DatabaseTypeName(),
		})
	}

	if !reflect.DeepEqual(actualColTypes, expectedColTypes) {
		t.Fatalf("cursor column not equal ,actual: \n%#v \n expected: \n%#v", actualColTypes, expectedColTypes)
	}

	for rows.Next() {
		// 获取单行中cursor列的数据
		cursor := sql.Rows{}
		if err := rows.Scan(&cursor); err != nil {
			t.Fatal(err)
		}

		cols, err := cursor.ColumnTypes()
		if err != nil {
			t.Fatal(err)
		}

		l := len(cols)

		// 校验cursor列的数据
		cursorData := make([][]any, 0)
		for cursor.Next() {
			data := make([]any, l)
			di := make([]any, l)
			for i := range data {
				di[i] = &data[i]
			}
			if err := cursor.Scan(di...); err != nil {
				t.Fatal(err)
			}
			cursorData = append(cursorData, data)
		}

		expected := [][]any{{int32(1), "9999"}}
		if !reflect.DeepEqual(cursorData, expected) {
			t.Fatalf("cursor data not equal ,actual: \n%#v \n expected: \n%#v", cursorData[0], expected)
		}
		_ = cursor.Close()

	}

	_ = rows.Close()
}

func execute(t *testing.T, conn *sql.Conn, sql string) {
	if _, err := conn.ExecContext(context.Background(), sql); err != nil {
		t.Fatal(err)
	}

}

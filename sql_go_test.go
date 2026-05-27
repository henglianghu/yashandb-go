// The following test cases are from https://github.com/bradfitz/go-sql-test
package yasdb

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestBlobGosql(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testblobgosql)
}

func testblobgosql(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	t.tableName = tablePrefix + "blob"
	t.columnNameType = [][2]string{
		{"id", IntegerPrimaryKey},
		{"bar", "blob"},
	}
	t.genTableTest()

	blob := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	t.mustExec(fmt.Sprintf("insert into %s (id,bar) values(?, ?)", t.tableName), 1, blob)

	want := fmt.Sprintf("%x", blob)
	b := make([]byte, 16)
	err := t.QueryRow(fmt.Sprintf("select bar from %s where id = ?", t.tableName), 1).Scan(&b)
	got := fmt.Sprintf("%x", b)
	if err != nil {
		t.Errorf("[]byte scan: %v", err)
	} else if got != want {
		t.Errorf("for []byte, got %q; want %q", got, want)
	}

	got = ""
	err = t.QueryRow(fmt.Sprintf("select bar from %s where id = ?", t.tableName), 1).Scan(&got)
	want = string(blob)
	if err != nil {
		t.Errorf("string scan: %v", err)
	} else if got != want {
		t.Errorf("for string, got %q; want %q", got, want)
	}
}

func TestClobGosql(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testclobgosql)
}

func testclobgosql(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	t.tableName = tablePrefix + "clob"
	t.columnNameType = [][2]string{
		{"id", IntegerPrimaryKey},
		{"bar", "clob"},
	}
	t.genTableTest()

	want := strings.Repeat("你好，YashanDB！", 10) + "......"
	t.mustExec(fmt.Sprintf("insert into %s (id,bar) values(?, ?)", t.tableName), 1, want)

	got := ""
	err := t.QueryRow(fmt.Sprintf("select bar from %s where id = ?", t.tableName), 1).Scan(&got)
	if err != nil {
		t.Errorf("string scan: %v", err)
	} else if got != want {
		t.Errorf("for string, got %q; want %q", got, want)
	}
}

func TestManyQueryRowGosql(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testmanyqueryrowgosql)
}

func testmanyqueryrowgosql(t *sqlTest) {
	if testing.Short() {
		t.Logf("it is short")
		return
	}
	t.sqlGenInfo = &sqlGenInfo{}
	t.tableName = tablePrefix + "MQR"
	t.columnNameType = [][2]string{
		{"id", IntegerPrimaryKey},
		{"name", "varchar(50)"},
	}
	t.genTableTest()

	t.mustExec(fmt.Sprintf("insert into %s (id, name) values(?,?)", t.tableName), 1, "ezreal")
	var name string
	total := 10000
	for i := 0; i < total; i++ {
		err := t.QueryRow(fmt.Sprintf("select name from %s where id = ?", t.tableName), 1).Scan(&name)
		if err != nil || name != "ezreal" {
			t.Fatalf("query row %d:%q failed, %v", i, name, err)
		}
	}
}

func TestTxQueryGosql(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testtxquerygosql)
}

func testtxquerygosql(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	t.tableName = tablePrefix + "txquery"
	t.columnNameType = [][2]string{
		{"id", IntegerPrimaryKey},
		{"name", "varchar(50)"},
	}
	t.genTableTest()

	tx, err := t.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	want := "你好，YahsanDB！"
	_, err = tx.Exec(fmt.Sprintf("insert into %s (id, name) values(?,?)", t.tableName), 1, want)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := tx.Query(fmt.Sprintf("select name from %s where id = ?", t.tableName), 1)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	if !rows.Next() {
		if rows.Err() != nil {
			t.Fatal(err)
		}
		t.Fatal("row next failed")
	}

	var got string
	err = rows.Scan(&got)
	if err != nil {
		t.Errorf("rows scan failed, %v", err)
	} else if got != want {
		t.Errorf("rows scan, but got no same %q != %q", got, want)
	}
}

func TestPrepareStmtGosql(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testpreparestmtgosql)
}

func testpreparestmtgosql(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	t.tableName = tablePrefix + "preparestmt"
	t.columnNameType = [][2]string{{"count", "int"}}
	t.genTableTest()

	selectStmt, err := t.Prepare(fmt.Sprintf("SELECT count FROM %s ORDER BY count DESC", t.tableName))
	if err != nil {
		t.Fatalf("select prepare failed, %v", err)
	}
	insertStmt, err := t.Prepare(fmt.Sprintf("INSERT INTO %s (count) VALUES (?)", t.tableName))
	if err != nil {
		t.Fatalf("insert prepare failed, %v", err)
	}

	for i := 1; i <= 3; i++ {
		if _, err := insertStmt.Exec(i); err != nil {
			t.Fatalf("execute %d failed, %v", i, err)
		}
	}

	total := 10
	queryChan := make(chan struct{})
	for x := 0; x < total; x++ {
		go func() {
			defer func() {
				queryChan <- struct{}{}
			}()
			for y := 0; y < 10; y++ {
				sum := 0
				if err := selectStmt.QueryRow().Scan(&sum); err != nil {
					if err != sql.ErrNoRows {
						t.Errorf("query %d failed, %v", y, err)
						return
					}

				}
				if _, err := insertStmt.Exec(rand.Intn(total * 10)); err != nil {
					t.Errorf("insert %d failed, %v", y, err)
					return
				}
			}
		}()
	}
	for i := 0; i < total; i++ {
		<-queryChan
	}
}

func TestEmojiGosqlTest(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testemojigosqltest)
}

func testemojigosqltest(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	t.tableName = tablePrefix + "emoji"
	t.columnNameType = [][2]string{
		{"id", IntegerPrimaryKey},
		{"c1", "varchar(10)"},
		{"c2", "varchar(10)"},
	}
	t.genTableTest()

	wantC1 := "😁"
	wantC2 := "😮"
	t.mustExec(fmt.Sprintf("insert into %s (id,c1,c2) values(?, ?, ?)", t.tableName), 1, wantC1, wantC2)

	gotC1 := ""
	gotC2 := ""
	err := t.QueryRow(fmt.Sprintf("select c1,c2 from %s where id = ?", t.tableName), 1).Scan(&gotC1, &gotC2)
	if err != nil {
		t.Errorf("[]byte scan: %v", err)
	} else if gotC1 != wantC1 || gotC2 != wantC2 {
		t.Errorf("for emoji, got %s, %s; want %s, %s", gotC1, gotC2, wantC1, wantC2)
	}
}

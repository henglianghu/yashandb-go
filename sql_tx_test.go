package yasdb

import (
	"context"
	"fmt"
	"testing"

	"github.com/yashan-technologies/yashandb-go/assert"
)

func TestTxCommit(t *testing.T) {
	db := newSqlTest(t)
	defer db.Close()
	db.mustExec("drop table if exists test_tx_commit")
	db.mustExec("create table test_tx_commit(c1 int,c2 varchar(20))")

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin tx err: %v", err)
	}
	_, err = tx.Exec("insert into test_tx_commit values(1,'000001')")
	if err != nil {
		t.Fatalf("exec tx err: %v", err)
	}
	_, err = tx.Exec("insert into test_tx_commit values(2,'000001')")
	if err != nil {
		t.Fatalf("exec tx err: %v", err)
	}
	_, err = tx.Query("select * from test_tx_commit")
	if err != nil {
		t.Fatalf("exec tx err: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit tx err: %v", err)
	}
}

func TestTxRollback(t *testing.T) {
	db := newSqlTest(t)
	defer db.Close()
	db.mustExec("drop table if exists test_tx_rollback")
	db.mustExec("create table test_tx_rollback(c1 int,c2 varchar(20))")

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin tx err: %v", err)
	}
	_, err = tx.Exec("insert into test_tx_rollback values(1,'000001')")
	if err != nil {
		t.Fatalf("exec tx err: %v", err)
	}
	_, err = tx.Exec("insert into test_tx_rollback values(2,'000001')")
	if err != nil {
		t.Fatalf("exec tx err: %v", err)
	}
	_, err = tx.Query("select * from test_tx_rollback")
	if err != nil {
		t.Fatalf("exec tx err: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback tx err: %v", err)
	}
}

func TestTxAutoCommit(t *testing.T) {
	// set autocommit true
	db := newSqlAutoCommitTest(t)
	defer db.Close()
	db.mustExec("drop table if exists t1")
	db.mustExec("create table t1(c1 int)")

	// start tx
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin tx err: %v", err)
	}
	_, err = tx.Exec("insert into t1 values(1)")
	if err != nil {
		t.Fatalf("exec tx err: %v", err)
	}
	_, err = tx.Exec("insert into t1 values(2)")
	if err != nil {
		t.Fatalf("exec tx err: %v", err)
	}
	_, err = tx.Query("select * from t1")
	if err != nil {
		t.Fatalf("exec tx err: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback tx err: %v", err)
	}
	var count int
	err = db.QueryRow("select count(*) from t1").Scan(&count)
	if err != nil {
		t.Fatalf("query after rollback tx err: %v", err)
	}
	if count != 0 {
		t.Errorf("after rollback expect count=0,in fact count=%d", count)
	}

}

func TestTxAutoCommit2(t *testing.T) {
	// set autocommit true
	db := newSqlAutoCommitTest(t)
	defer db.Close()
	db.mustExec("drop table if exists t1")
	db.mustExec("create table t1(c1 int)")

	// start tx
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin tx err: %v", err)
	}
	_, err = tx.Exec("insert into t1 values(1)")
	if err != nil {
		t.Fatalf("exec tx err: %v", err)
	}
	_, err = tx.Exec("insert into t1 values(2)")
	if err != nil {
		t.Fatalf("exec tx err: %v", err)
	}
	_, err = tx.Query("select * from t1")
	if err != nil {
		t.Fatalf("exec tx err: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit tx err: %v", err)
	}
	var count int
	err = db.QueryRow("select count(*) from t1").Scan(&count)
	if err != nil {
		t.Fatalf("query after commit tx err: %v", err)
	}
	if count != 2 {
		t.Errorf("after commit expect count=2,in fact count=%d", count)
	}

	// test auto commit open now
	_, err = db.Exec("insert into t1 values (3)")
	if err != nil {
		t.Fatalf("query err: %v", err)
	}

	db2 := newSqlAutoCommitTest(t)
	err = db2.QueryRow("select count(*) from t1").Scan(&count)
	if err != nil {
		t.Fatalf("query after commit tx err: %v", err)
	}
	if count != 3 {
		t.Errorf("after commit expect count=3,in fact count=%d", count)
	}
}

func TestYasStmt_NumInput(t *testing.T) {

	db := newSqlTest(t)
	defer db.Close()

	db.mustExec("drop table if exists users")
	db.mustExec("create table users(id int, name varchar(20), age int)")
	conn := db.Driver().(*YasdbDriver).Conn()

	at := assert.NewAssert(t)

	// 测试用例表格
	tests := []struct {
		name    string
		prepare func() (*YasStmt, error)
		want    int
		wantErr string
	}{
		{
			name: "stmt为nil时返回-1",
			prepare: func() (*YasStmt, error) {
				return &YasStmt{Stmt: nil}, nil
			},
			want:    -1,
			wantErr: "",
		},
		{
			name: "解析SQL参数失败时返回-1",
			prepare: func() (*YasStmt, error) {
				return createTestStmt(conn, "INVALID SQL SYNTAX!")
			},
			want:    -1,
			wantErr: "YAS-04231",
		}, {
			name: "无参数SQL返回0",
			prepare: func() (*YasStmt, error) {
				return createTestStmt(conn, "SELECT * FROM users")
			},
			want: 0,
		}, {
			name: "单参数SQL返回1",
			prepare: func() (*YasStmt, error) {
				return createTestStmt(conn, "SELECT * FROM users WHERE id = ?")
			},
			want: 1,
		}, {
			name: "单名称参数SQL返回1",
			prepare: func() (*YasStmt, error) {
				return createTestStmt(conn, "SELECT * FROM users WHERE id = :id")
			},
			want: 1,
		}, {
			name: "多参数SQL返回对应数量",
			prepare: func() (*YasStmt, error) {
				return createTestStmt(conn, "INSERT INTO users (id, name, age) VALUES (?, ?, ?)")
			},
			want: 3,
		}, {
			name: "多名称参数SQL返回对应数量",
			prepare: func() (*YasStmt, error) {
				return createTestStmt(conn, "INSERT INTO users (id, name, age) VALUES (:id, :name, :age)")
			},
			want: 3,
		},
	}

	// 执行测试
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(tt.name)
			stmt, err := tt.prepare()
			if err != nil {
				at.Contains(err.Error(), tt.wantErr)
				return
			}
			num := stmt.NumInput()
			fmt.Printf("sql: %s, want: %d, real: %d\n", stmt.Sqlstr, tt.want, num)
			at.Equal(tt.want, num)
		})
	}
}

// 测试辅助函数：创建测试用YasStmt实例
func createTestStmt(conn *YasConn, sql string) (*YasStmt, error) {
	stmt, err := PrepareContext(conn, context.Background(), sql)
	if err != nil {
		return nil, err
	}
	return stmt, nil
}

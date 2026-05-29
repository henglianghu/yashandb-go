package yasdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"testing"

	"github.com/yashan-technologies/yashandb-go/assert"
)

func TestConnect(t *testing.T) {
	db, err := sql.Open("yasdb", testDsn)
	if err != nil {
		t.Fatalf("%s%v", NormalConnErr, err)
	}
	if err = db.Close(); err != nil {
		t.Fatalf("error db close: %v", err)
	}
}

/*
must set env:
LD_LIBRARY_PATH
YASDB_DATA
*/
func TestConnectWithoutPassword(t *testing.T) {
	db, err := sql.Open("yasdb", "")
	if err != nil {
		t.Fatalf("%s%v", NormalConnErr, err)
	}
	if err = db.Close(); err != nil {
		t.Fatalf("error db close: %v", err)
	}
}

func TestPing(t *testing.T) {
	db, err := sql.Open("yasdb", testDsn)
	if err != nil {
		t.Fatalf("%s%v", NormalConnErr, err)
	}
	if err = db.Ping(); err != nil {
		t.Fatalf("error db ping: %v", err)
	}
	if err = db.Close(); err != nil {
		t.Fatalf("error db close: %v", err)
	}
	if err = db.Ping(); err == nil || !strings.Contains(err.Error(), "database is closed") {
		t.Fatalf("close db but ping success: %v", err)
	}
}

func TestAutoCommitTrue(t *testing.T) {
	runsqlTestACTrue(t, testAutoCommitTrue)
}

func testAutoCommitTrue(t *sqlTest) {
	si := sqlGenInfo{}
	t.sqlGenInfo = &si

	si = sqlGenInfo{
		tableName: "test_auto_commit_true",
		columnNameType: [][2]string{
			{"id", "int"},
			{"name", "varchar(20)"},
		},
		execArgs: [][]interface{}{
			{1, "column1"},
			{2, "column2"},
			{3, "column3"},
		},
		queryResult: [][]interface{}{
			{int32(1), "column1"},
			{int32(2), "column2"},
			{int32(3), "column3"},
		},
	}
	t.genTableTest()
	t.runInsertTest()
	t.runSelectTest()
}

func TestAutoCommitFase(t *testing.T) {
	runsqlTestACFalse(t, testAutoCommitFalse)
}

func testAutoCommitFalse(t *sqlTest) {
	si := sqlGenInfo{}
	t.sqlGenInfo = &si

	si = sqlGenInfo{
		tableName: "test_auto_commit_false",
		columnNameType: [][2]string{
			{"id", "int"},
			{"name", "varchar(20)"},
		},
		execArgs: [][]interface{}{
			{1, "column1"},
			{2, "column2"},
			{3, "column3"},
		},
		queryResult: [][]interface{}{
			{int32(1), "column1"},
			{int32(2), "column2"},
			{int32(3), "column3"},
		},
	}
	t.genTableTest()
	t.runInsertTest()
	t.runSelectTest()
}

func getYasConn(t *testing.T) *YasConn {
	db, err := sql.Open("yasdb", testDsn)
	if err != nil {
		t.Fatalf("%s%v", NormalConnErr, err)
	}
	if err = db.Ping(); err != nil {
		t.Fatalf("error db ping: %v", err)
	}
	hConn := db.Driver().(*YasdbDriver).Conn()
	return hConn
}

func TestYasConn_ResetSession(t *testing.T) {

	at := assert.NewAssert(t)

	// 定义测试用例
	tests := []struct {
		name    string
		prepare func() (*YasConn, context.Context)
		wantErr error
	}{{
		name: "上下文已取消",
		prepare: func() (*YasConn, context.Context) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // 立即取消上下文
			return getYasConn(t), ctx
		},
		wantErr: context.Canceled,
	}, {
		name: "连接为nil",
		prepare: func() (*YasConn, context.Context) {
			return nil, context.Background()
		},
		wantErr: driver.ErrBadConn,
	}, {
		name: "正常状态",
		prepare: func() (*YasConn, context.Context) {
			hConn := getYasConn(t)
			_, err := hConn.Prepare("select 1 from dual")
			if err != nil {
				t.Fatalf("error db exec: %v", err)
			}
			return hConn, context.Background()
		},
		wantErr: nil,
	}, {
		name: "连接已关闭",
		prepare: func() (*YasConn, context.Context) {
			hConn := getYasConn(t)
			hConn.closed = true
			return hConn, context.Background()
		},
		wantErr: driver.ErrBadConn,
	}, {
		name: "服务器状态异常",
		prepare: func() (*YasConn, context.Context) {
			hConn := getYasConn(t)
			hConn.Close()
			hConn.closed = false
			return hConn, context.Background()
		},
		wantErr: driver.ErrBadConn,
	}}

	// 执行测试
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(tt.name, "begin")
			conn, ctx := tt.prepare()
			// 调用待测试方法
			var err error
			if conn == nil {
				// 特殊处理nil连接情况
				var connPtr *YasConn
				err = connPtr.ResetSession(ctx)
			} else {
				err = conn.ResetSession(ctx)
			}
			fmt.Println(tt.name, err)
			// 验证结果
			at.Equal(tt.wantErr, err)
		})
	}
}

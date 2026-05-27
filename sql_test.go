package yasdb

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestBindIntputParam(t *testing.T) {
	runSqlTest(t, testBindIntputParam)
}

func testBindIntputParam(t *sqlTest) {
	si := sqlGenInfo{}
	t.sqlGenInfo = &si

	si = sqlGenInfo{
		tableName: "test_bind_intput_param",
		columnNameType: [][2]string{
			{"c1", "int"},
			{"c2", "double"},
			{"c3", "number"},
			{"c4", "date"},
			{"c5", "timestamp"},
			{"c6", "char(20)"},
			{"c7", "varchar(20)"},
			{"c8", "clob"},
			{"c9", "blob"},
			{"c10", "boolean"},
		},
	}
	t.genTableTest()
	t.mustExec(
		fmt.Sprintf("insert into %s(c1, c2, c3, c4, c5, c6, c7, c8, c9, c10) values(?,?,?,?,?,?,?,?,?,?)", si.tableName),
		1, 2.2, 3.3, "2022-10-08", "2022-10-08 16:01:01", "你好，YashanDB！", "你好，YashanDB！", "你好，YashanDB！", []byte("你好，YashanDB！"), true,
	)
	t.mustExec(
		fmt.Sprintf("insert into %s(c1, c2, c3, c4, c5, c6, c7, c8, c9, c10) values(:1, :2, :3, :4, :5, :6, :7, :8, :9, :10)", si.tableName),
		2, 3.2, 4.3, time.Now(), time.Now().AddDate(0, -1, -1), "c6", "c7", "c8", []byte("c9"), true,
	)
	t.mustQuery(fmt.Sprintf("select * from %s where c1 > ?", si.tableName), 1)
	t.mustQuery(fmt.Sprintf("select * from %s where c10 = ?", si.tableName), true)
}

func TestBindOutputParam(t *testing.T) {
	runSqlTest(t, testBindOutputParam)
}

func testBindOutputParam(t *sqlTest) {
	si := sqlGenInfo{}
	t.sqlGenInfo = &si

	si = sqlGenInfo{
		tableName: "test_bind_output_param",
		columnNameType: [][2]string{
			{"c1", "int"},
			{"c2", "double"},
			{"c3", "number"},
			{"c4", "date"},
			{"c5", "timestamp"},
			{"c6", "char(20)"},
			{"c7", "varchar(20)"},
			{"c8", "clob"},
			{"c9", "blob"},
			{"c10", "boolean"},
		},
	}
	t.genTableTest()
	var (
		c1  = 1
		c2  = 2.2
		c3  = 3.3
		c4  = time.Date(2022, 10, 8, 01, 01, 0, 0, time.UTC)
		c5  = time.Date(2022, 10, 8, 01, 01, 0, 0, time.UTC).AddDate(-1, -1, -1)
		c6  = "你好，YashanDB！"
		c7  = strings.Repeat("c", 20)
		c8  = c6
		c9  = []byte(c8)
		c10 = true
	)

	var (
		outC1  int
		outC2  float64
		outC3  float64
		outC4  time.Time
		outC5  time.Time
		outC6  string
		outC7  string
		outC8  string
		outC9  []byte
		outC10 bool
	)
	outBindC6, _ := NewOutputBindValue(&outC6, WithTypeChar(), WithBindSize(21))
	outBindC7, _ := NewOutputBindValue(&outC7, WithTypeVarchar(), WithBindSize(21))
	outBindC8, _ := NewOutputBindValue(&outC8, WithTypeClob())
	outBindC9, _ := NewOutputBindValue(&outC9, WithTypeBlob())

	t.mustExec(
		fmt.Sprintf(`insert into %s(c1, c2, c3, c4, c5, c6, c7, c8, c9, c10) values(?,?,?,?,?,?,?,?,?,?)
        returning c1, c2, c3, c4, c5, c6, c7, c8, c9, c10 into ?,?,?,?,?,?,?,?,?,?`, si.tableName),
		c1,
		c2,
		c3,
		c4,
		c5,
		c6,
		c7,
		c8,
		c9,
		c10,
		sql.Out{Dest: &outC1},
		sql.Out{Dest: &outC2},
		sql.Out{Dest: &outC3},
		sql.Out{Dest: &outC4},
		sql.Out{Dest: &outC5},
		sql.Out{Dest: outBindC6},
		sql.Out{Dest: outBindC7},
		sql.Out{Dest: outBindC8},
		sql.Out{Dest: outBindC9},
		sql.Out{Dest: &outC10},
	)
	if c1 != outC1 ||
		c2 != outC2 ||
		c3 != outC3 ||
		c4.Unix() != outC4.Unix() ||
		c5.UnixMilli() != outC5.UnixMilli() ||
		c6 != outC6 ||
		c7 != outC7 ||
		c8 != outC8 ||
		string(c9) != string(outC9) ||
		c10 != outC10 {
		t.Fatalf("output param value is wrong!")
	}
}

func TestCreateTable(t *testing.T) {
	db := newSqlTest(t)
	defer db.Close()
	db.mustExec("drop table if exists test_create_table1")
	db.mustExec("create table test_create_table1(c1 int)")

	db.mustExec("drop table if exists test_create_table2")
	db.mustExec(`create table test_create_table2(
                area_no CHAR(2) NOT NULL,
                area_name VARCHAR2(60),
                DHQ VARCHAR2(20) DEFAULT 'ShenZhen' NOT NULL)`)

	db.mustExec("drop table if exists test_create_table3")
	db.mustExec(`create table test_create_table3(
        branch_no CHAR(4) PRIMARY KEY,
        branch_name VARCHAR2(200) NOT NULL,
        area_no CHAR(2),
        address VARCHAR2(200))`)

	db.mustExec("drop table if exists test_create_table4")
	db.mustExec(`create table test_create_table4(
        product_no CHAR(5) PRIMARY KEY,
        product_name VARCHAR2(30),
        cost NUMBER,
        price NUMBER)`)
}

func TestDropTable(t *testing.T) {
	db := newSqlTest(t)
	defer db.Close()
	db.mustExec("drop table if exists test_drop_table1")
	db.mustExec("create table test_drop_table1(c1 int)")
	db.mustExec("drop table test_drop_table1")

	db.mustExec("drop table if exists test_drop_table2")
	db.mustExec(`create table test_drop_table2(
                area_no CHAR(2) NOT NULL,
                area_name VARCHAR2(60),
                DHQ VARCHAR2(20) DEFAULT 'ShenZhen' NOT NULL)`)
	db.mustExec("drop table test_drop_table2")
}

func TestAlterTable(t *testing.T) {
	db := newSqlTest(t)
	defer db.Close()
	db.mustExec("drop table if exists test_alter_table1")
	db.mustExec("create table test_alter_table1(c1 int)")
	db.mustExec("alter table test_alter_table1 add(c2 varchar(20))")
	db.mustExec("alter table test_alter_table1 add(c3 float)")
	db.mustExec("alter table test_alter_table1 drop(c1)")
}

func TestSelectFromDual(t *testing.T) {
	db := newSqlTest(t)
	defer db.Close()

	db.mustQuery("select 1 from dual")
	db.mustQuery("select '😀' from dual")
}

func TestSelectFromTable(t *testing.T) {
	db := newSqlTest(t)
	defer db.Close()

	db.mustExec("drop table if exists test_select1")
	db.mustExec(`CREATE TABLE test_select1(
        branch_no CHAR(4) PRIMARY KEY,
        branch_name VARCHAR2(200) NOT NULL,
        area_no CHAR(2),
        address VARCHAR2(200))`)
	db.mustExec("INSERT INTO test_select1 VALUES ('0001','深圳','','')")
	db.mustExec("INSERT INTO test_select1 VALUES ('0101','上海','01','上海市静安区')")
	db.mustExec("INSERT INTO test_select1 VALUES ('0102','南京','01','City of Nanjing')")
	db.mustExec("INSERT INTO test_select1 VALUES ('0103','福州','01','')")
	db.mustExec("INSERT INTO test_select1 VALUES ('0104','厦门','01','Xiamen')")
	db.mustExec("INSERT INTO test_select1 VALUES ('0401','北京','04','')")
	db.mustExec("INSERT INTO test_select1 VALUES ('0402','天津','04','')")
	db.mustExec("INSERT INTO test_select1 VALUES ('0403','大连','04','大连市')")
	db.mustExec("INSERT INTO test_select1 VALUES ('0404','沈阳','04','')")
	db.mustExec("INSERT INTO test_select1 VALUES ('0201','成都','02','')")
	db.mustExec("INSERT INTO test_select1 VALUES ('0501','武汉','','')")
	db.mustExec("INSERT INTO test_select1 VALUES ('0502','长沙','05','')")

	db.mustQuery("SELECT * FROM test_select1 LIMIT 4")
	db.mustQuery("SELECT * FROM test_select1 LIMIT 4 OFFSET 3")
}

func TestInsert(t *testing.T) {
	db := newSqlTest(t)
	defer db.Close()

	db.mustExec("drop table if exists test_insert1")
	db.mustExec(`CREATE TABLE test_insert1(
        branch_no CHAR(4) PRIMARY KEY,
        branch_name VARCHAR2(200) NOT NULL,
        area_no CHAR(2),
        address VARCHAR2(200))`)
	db.mustExec("INSERT INTO test_insert1 VALUES ('0001','深圳','','')")
	db.mustExec("INSERT INTO test_insert1 VALUES ('0101','上海','01','上海市静安区')")
	db.mustExec("INSERT INTO test_insert1 VALUES ('0102','南京','01','City of Nanjing')")
	db.mustExec("INSERT INTO test_insert1 VALUES ('0103','福州','01','')")
	db.mustExec("INSERT INTO test_insert1 VALUES (?,?,?,?)", "0403", "大连", "04", "大连市")
	db.mustExec("INSERT INTO test_insert1 VALUES (?,?,?,?)", "0404", "沈阳", "04", "")
	db.mustExec("INSERT INTO test_insert1 VALUES (:1,:2,:3,:4)", "0201", "成都", "02", "")
	db.mustExec("INSERT INTO test_insert1 VALUES (:1,:2,:3,:4)", "0501", "武汉", "", "")
}

func TestUpdate(t *testing.T) {
	db := newSqlTest(t)
	defer db.Close()

	db.mustExec("drop table if exists test_update1")
	db.mustExec(`CREATE TABLE test_update1(
        order_no CHAR(14) NOT NULL,
        product_no CHAR(5) ,
        area CHAR(2) ,
        branch CHAR(4) ,
        order_date DATE DEFAULT SYSDATE NOT NULL,
        salesperson CHAR(10) ,
        id NUMBER)`)
	db.mustExec("INSERT INTO test_update1 VALUES ('20010102020001','11001','02','0201',sysdate-400,'0201010011',300)")
	db.mustExec("INSERT INTO test_update1 VALUES ('20010102020002','11002','02','0201',sysdate-400,'0201008003',1300)")
	db.mustExec("INSERT INTO test_update1 VALUES ('20010102020003','10001','02','0201',sysdate-400,'0201010011',2300)")
	db.mustExec("INSERT INTO test_update1 VALUES ('20210102020004','11001','02','0201',sysdate-400,'0201008003',400)")
	db.mustExec("INSERT INTO test_update1 VALUES ('20210102020005','11002','02','0201',sysdate-400,'0201010011',200)")
	db.mustExec("INSERT INTO test_update1 VALUES ('20210102020006','10001','02','0201',sysdate-400,'0201008003',100)")
	db.mustExec("UPDATE test_update1 set product_no='11003' WHERE product_no='10001'")
	db.mustExec("UPDATE test_update1 set order_no='00001' WHERE id=100")
}

func TestDelete(t *testing.T) {
	db := newSqlTest(t)
	defer db.Close()

	db.mustExec("drop table if exists test_delete1")
	db.mustExec(`CREATE TABLE test_delete1(
        order_no CHAR(14) NOT NULL,
        product_no CHAR(5) ,
        area CHAR(2) ,
        branch CHAR(4) ,
        order_date DATE DEFAULT SYSDATE NOT NULL,
        salesperson CHAR(10) ,
        id NUMBER)`)
	db.mustExec("INSERT INTO test_delete1 VALUES ('20010102020001','11001','02','0201',sysdate-400,'0201010011',300)")
	db.mustExec("INSERT INTO test_delete1 VALUES ('20010102020002','11002','02','0201',sysdate-400,'0201008003',1300)")
	db.mustExec("INSERT INTO test_delete1 VALUES ('20010102020003','10001','02','0201',sysdate-400,'0201010011',2300)")
	db.mustExec("INSERT INTO test_delete1 VALUES ('20210102020004','11001','02','0201',sysdate-400,'0201008003',400)")
	db.mustExec("INSERT INTO test_delete1 VALUES ('20210102020005','11002','02','0201',sysdate-400,'0201010011',200)")
	db.mustExec("INSERT INTO test_delete1 VALUES ('20210102020006','10001','02','0201',sysdate-400,'0201008003',100)")
	db.mustExec("DELETE FROM test_delete1 WHERE order_no=20010102020002")
	db.mustExec("DELETE FROM test_delete1 WHERE id<300")
}

func TestDDLResult(t *testing.T) {
	db := newSqlTest(t)
	defer db.Close()

	result := db.mustExec("drop table if exists test_ddl_result")
	affectedRows, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("get RowsAffected err: %v", err)
	}
	affectedResultComparison(t, affectedRows, 0)

	result = db.mustExec("create table test_ddl_result(c1 int)")
	affectedRows, err = result.RowsAffected()
	if err != nil {
		t.Fatalf("get RowsAffected err: %v", err)
	}
	affectedResultComparison(t, affectedRows, 0)

	result = db.mustExec("alter table test_ddl_result add(c2 int)")
	affectedRows, err = result.RowsAffected()
	if err != nil {
		t.Fatalf("get RowsAffected err: %v", err)
	}
	affectedResultComparison(t, affectedRows, 0)
}

func TestDMLResult(t *testing.T) {
	db := newSqlTest(t)
	defer db.Close()
	db.mustExec("drop table if exists test_dml_result")
	db.mustExec("create table test_dml_result(c1 int)")

	result := db.mustExec("insert into test_dml_result (c1) values (1)")
	affectedRows, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("get RowsAffected err: %v", err)
	}
	affectedResultComparison(t, affectedRows, 1)

	result = db.mustExec("insert into test_dml_result (c1) values (2)")
	affectedRows, err = result.RowsAffected()
	if err != nil {
		t.Fatalf("get RowsAffected err: %v", err)
	}
	affectedResultComparison(t, affectedRows, 1)

	result = db.mustExec("select * from test_dml_result")
	_, err = result.RowsAffected()
	if err != nil {
		t.Fatalf("get RowsAffected err: %v", err)
	}
	// affectedResultComparison(t, affectedRows, 2)

	result = db.mustExec("delete from test_dml_result where c1=2")
	_, err = result.RowsAffected()
	if err != nil {
		t.Fatalf("get RowsAffected err: %v", err)
	}
	// affectedResultComparison(t, affectedRows, 1)

	result = db.mustExec("update test_dml_result set c1=10000 where c1=1")
	_, err = result.RowsAffected()
	if err != nil {
		t.Fatalf("get RowsAffected err: %v", err)
	}
	// affectedResultComparison(t, affectedRows, 1)
}

func TestQueryContainSemicolon(t *testing.T) {
	db := newSqlTest(t)
	defer db.Close()

	db.mustExec("drop table if exists test_semicolon;")
	db.mustExec(`CREATE TABLE test_semicolon(
        order_no CHAR(14) NOT NULL,
        product_no CHAR(5) ,
        area CHAR(2) ,
        branch CHAR(4) ,
        order_date DATE DEFAULT SYSDATE NOT NULL,
        salesperson CHAR(10) ,
        id NUMBER);  `)
	db.mustExec("INSERT INTO test_semicolon VALUES ('20010102020001','11001','02','0201',sysdate-400,'0201010011',300);")
	db.mustExec("INSERT INTO test_semicolon VALUES ('20010102020002','11002','02','0201',sysdate-400,'0201008003',1300) ; ")
	db.mustExec("INSERT INTO test_semicolon VALUES ('20010102020003','10001','02','0201',sysdate-400,'0201010011',2300) ;  ")
	db.mustExec("INSERT INTO test_semicolon VALUES ('20210102020004','11001','02','0201',sysdate-400,'0201008003',400)  ;")
	db.mustExec("INSERT INTO test_semicolon VALUES ('20210102020005','11002','02','0201',sysdate-400,'0201010011',200);")
	db.mustExec("INSERT INTO test_semicolon VALUES ('20210102020006','10001','02','0201',sysdate-400,'0201008003',100)   ;")
	db.mustQuery("select * from test_semicolon;")
	db.mustExec("DELETE FROM test_semicolon WHERE order_no=20010102020002; ")
	db.mustExec("DELETE FROM test_semicolon WHERE id<300  ;")
}

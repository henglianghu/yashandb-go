package yasdb

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestOutputBindByDest_ReturningInto(t *testing.T) {
	runSqlTest(t, testOutputBindByDestReturningInto)
}

func testOutputBindByDestReturningInto(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	t.tableName = "test_output_bind_dest"
	t.columnNameType = [][2]string{
		{"c1", "bigint"},
		{"c2", "double"},
		{"c3", "varchar(100)"},
		{"c4", "timestamp"},
		{"c5", "boolean"},
	}
	t.genTableTest()
	defer t.dropTable()

	var (
		inC1 = int64(9223372036854775807)
		inC2 = float64(3.14159265358979)
		inC3 = "hello yashandb"
		inC4 = time.Date(2024, 6, 15, 10, 30, 45, 0, time.UTC)
		inC5 = true
	)

	var (
		outC1 int64
		outC2 float64
		outC3 string
		outC4 time.Time
		outC5 bool
	)

	t.mustExec(
		fmt.Sprintf(`insert into %s(c1, c2, c3, c4, c5) values(?,?,?,?,?)
		returning c1, c2, c3, c4, c5 into ?,?,?,?,?`, t.tableName),
		inC1, inC2, inC3, inC4, inC5,
		sql.Out{Dest: &outC1},
		sql.Out{Dest: &outC2},
		sql.Out{Dest: &outC3},
		sql.Out{Dest: &outC4},
		sql.Out{Dest: &outC5},
	)

	if outC1 != inC1 {
		t.Fatalf("c1(bigint): got %d, want %d", outC1, inC1)
	}
	if outC2 != inC2 {
		t.Fatalf("c2(double): got %f, want %f", outC2, inC2)
	}
	if outC3 != inC3 {
		t.Fatalf("c3(varchar): got %q, want %q", outC3, inC3)
	}
	if outC4.Unix() != inC4.Unix() {
		t.Fatalf("c4(timestamp): got %v, want %v", outC4, inC4)
	}
	if outC5 != inC5 {
		t.Fatalf("c5(boolean): got %v, want %v", outC5, inC5)
	}
}

func TestOutputBindByDest_BigintMaxMin(t *testing.T) {
	runSqlTest(t, testOutputBindByDestBigintMaxMin)
}

func testOutputBindByDestBigintMaxMin(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	t.tableName = "test_output_bigint_range"
	t.columnNameType = [][2]string{
		{"c1", "bigint"},
	}
	t.genTableTest()
	defer t.dropTable()

	cases := []struct {
		name string
		val  int64
	}{
		{"max", int64(9223372036854775807)},
		{"min", int64(-9223372036854775808)},
		{"zero", int64(0)},
		{"negative", int64(-1)},
		{"overflow_int32_max", int64(2147483648)},
		{"overflow_int32_min", int64(-2147483649)},
	}

	for _, c := range cases {
		var out int64
		t.mustExec(
			fmt.Sprintf(`insert into %s(c1) values(?) returning c1 into ?`, t.tableName),
			c.val,
			sql.Out{Dest: &out},
		)
		if out != c.val {
			t.Fatalf("case %s: got %d, want %d", c.name, out, c.val)
		}
	}
}

func TestOutputBindByDest_Float64Precision(t *testing.T) {
	runSqlTest(t, testOutputBindByDestFloat64Precision)
}

func testOutputBindByDestFloat64Precision(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	t.tableName = "test_output_float64_prec"
	t.columnNameType = [][2]string{
		{"c1", "double"},
	}
	t.genTableTest()
	defer t.dropTable()

	cases := []struct {
		name string
		val  float64
	}{
		{"positive", 1234567890.123456789},
		{"negative", -9876543210.987654321},
		{"zero", 0.0},
		{"small", 0.000000001},
	}

	for _, c := range cases {
		var out float64
		t.mustExec(
			fmt.Sprintf(`insert into %s(c1) values(?) returning c1 into ?`, t.tableName),
			c.val,
			sql.Out{Dest: &out},
		)
		if out != c.val {
			t.Fatalf("case %s: got %v, want %v", c.name, out, c.val)
		}
	}
}

func TestOutputBindByDest_BoolValues(t *testing.T) {
	runSqlTest(t, testOutputBindByDestBoolValues)
}

func testOutputBindByDestBoolValues(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	t.tableName = "test_output_bool"
	t.columnNameType = [][2]string{
		{"c1", "boolean"},
	}
	t.genTableTest()
	defer t.dropTable()

	for _, val := range []bool{true, false} {
		var out bool
		t.mustExec(
			fmt.Sprintf(`insert into %s(c1) values(?) returning c1 into ?`, t.tableName),
			val,
			sql.Out{Dest: &out},
		)
		if out != val {
			t.Fatalf("bool: got %v, want %v", out, val)
		}
	}
}

func TestOutputBindByDest_StringValues(t *testing.T) {
	runSqlTest(t, testOutputBindByDestStringValues)
}

func testOutputBindByDestStringValues(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	t.tableName = "test_output_string"
	t.columnNameType = [][2]string{
		{"c1", "varchar(32000)"},
	}
	t.genTableTest()
	defer t.dropTable()

	cases := []struct {
		name string
		val  string
	}{
		{"ascii", "hello yashandb"},
		{"chinese", "你好，崖山数据库！"},
		{"empty", ""},
		{"long", strings.Repeat("abcdefghij", 1000)},
	}

	for _, c := range cases {
		var out string
		t.mustExec(
			fmt.Sprintf(`insert into %s(c1) values(?) returning c1 into ?`, t.tableName),
			c.val,
			sql.Out{Dest: &out},
		)
		if out != c.val {
			t.Fatalf("case %s: got len=%d, want len=%d", c.name, len(out), len(c.val))
		}
	}
}

func TestOutputBindByDest_TimestampValues(t *testing.T) {
	runSqlTest(t, testOutputBindByDestTimestampValues)
}

func testOutputBindByDestTimestampValues(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	t.tableName = "test_output_timestamp"
	t.columnNameType = [][2]string{
		{"c1", "timestamp"},
	}
	t.genTableTest()
	defer t.dropTable()

	cases := []struct {
		name string
		val  time.Time
	}{
		{"normal", time.Date(2024, 6, 15, 10, 30, 45, 0, time.UTC)},
		{"epoch", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"future", time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)},
	}

	for _, c := range cases {
		var out time.Time
		t.mustExec(
			fmt.Sprintf(`insert into %s(c1) values(?) returning c1 into ?`, t.tableName),
			c.val,
			sql.Out{Dest: &out},
		)
		if out.Unix() != c.val.Unix() {
			t.Fatalf("case %s: got %v, want %v", c.name, out, c.val)
		}
	}
}

func TestOutputBindByDest_ProcedureCall(t *testing.T) {
	runSqlTest(t, testOutputBindByDestProcedureCall)
}

func testOutputBindByDestProcedureCall(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}

	proc := `
	CREATE OR REPLACE PROCEDURE p_test_dest_out(
		p_int OUT BIGINT,
		p_dbl OUT DOUBLE,
		p_str OUT VARCHAR,
		p_bool OUT BOOLEAN
	) AS
	BEGIN
		p_int := 9223372036854775807;
		p_dbl := 2.718281828;
		p_str := 'from procedure';
		p_bool := TRUE;
	END;
	`
	t.mustExec(proc)

	var (
		outInt  int64
		outDbl  float64
		outStr  string
		outBool bool
	)

	t.mustExec(
		`BEGIN "P_TEST_DEST_OUT"(:1, :2, :3, :4); END;`,
		sql.Out{Dest: &outInt},
		sql.Out{Dest: &outDbl},
		sql.Out{Dest: &outStr},
		sql.Out{Dest: &outBool},
	)

	if outInt != 9223372036854775807 {
		t.Fatalf("int: got %d, want 9223372036854775807", outInt)
	}
	if outDbl != 2.718281828 {
		t.Fatalf("double: got %f, want 2.718281828", outDbl)
	}
	if outStr != "from procedure" {
		t.Fatalf("string: got %q, want %q", outStr, "from procedure")
	}
	if outBool != true {
		t.Fatalf("bool: got %v, want true", outBool)
	}
}

func TestOutputBindByDest_ProcedureInout(t *testing.T) {
	runSqlTest(t, testOutputBindByDestProcedureInout)
}

func testOutputBindByDestProcedureInout(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}

	proc := `
	CREATE OR REPLACE PROCEDURE p_test_dest_inout(
		p_int IN OUT BIGINT,
		p_dbl IN OUT DOUBLE,
		p_str IN OUT VARCHAR,
		p_bool IN OUT BOOLEAN
	) AS
	BEGIN
		p_int := p_int + 1;
		p_dbl := p_dbl * 2;
		p_str := p_str || ' world';
		p_bool := NOT p_bool;
	END;
	`
	t.mustExec(proc)

	var (
		ioInt  = int64(100)
		ioDbl  = float64(1.5)
		ioStr  = "hello"
		ioBool = false
	)

	t.mustExec(
		`BEGIN "P_TEST_DEST_INOUT"(:1, :2, :3, :4); END;`,
		sql.Out{Dest: &ioInt, In: true},
		sql.Out{Dest: &ioDbl, In: true},
		sql.Out{Dest: &ioStr, In: true},
		sql.Out{Dest: &ioBool, In: true},
	)

	if ioInt != 101 {
		t.Fatalf("int inout: got %d, want 101", ioInt)
	}
	if ioDbl != 3.0 {
		t.Fatalf("double inout: got %f, want 3.0", ioDbl)
	}
	if ioStr != "hello world" {
		t.Fatalf("string inout: got %q, want %q", ioStr, "hello world")
	}
	if ioBool != true {
		t.Fatalf("bool inout: got %v, want true", ioBool)
	}
}

func TestOutputBindByDest_ProcedureTimestamp(t *testing.T) {
	runSqlTest(t, testOutputBindByDestProcedureTimestamp)
}

func testOutputBindByDestProcedureTimestamp(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}

	proc := `
	CREATE OR REPLACE PROCEDURE p_test_ts_dest(
		p_ts OUT TIMESTAMP
	) AS
	BEGIN
		p_ts := TO_TIMESTAMP('2024-12-25 08:30:00', 'YYYY-MM-DD HH24:MI:SS');
	END;
	`
	t.mustExec(proc)

	var outTs time.Time
	t.mustExec(
		`BEGIN "P_TEST_TS_DEST"(:1); END;`,
		sql.Out{Dest: &outTs},
	)

	expected, _ := time.Parse("2006-01-02 15:04:05", "2024-12-25 08:30:00")
	if outTs.Unix() != expected.Unix() {
		t.Fatalf("timestamp: got %v, want %v", outTs, expected)
	}
}

func TestOutputBindByDest_ProcedureTimestampInout(t *testing.T) {
	runSqlTest(t, testOutputBindByDestProcedureTimestampInout)
}

func testOutputBindByDestProcedureTimestampInout(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}

	proc := `
	CREATE OR REPLACE PROCEDURE p_test_ts_inout_dest(
		p_ts IN OUT TIMESTAMP
	) AS
	BEGIN
		p_ts := TO_TIMESTAMP('2025-01-01 00:00:00', 'YYYY-MM-DD HH24:MI:SS');
	END;
	`
	t.mustExec(proc)

	ioTs := time.Now()
	t.mustExec(
		`BEGIN "P_TEST_TS_INOUT_DEST"(:1); END;`,
		sql.Out{Dest: &ioTs, In: true},
	)

	expected, _ := time.Parse("2006-01-02 15:04:05", "2025-01-01 00:00:00")
	if ioTs.Unix() != expected.Unix() {
		t.Fatalf("timestamp inout: got %v, want %v", ioTs, expected)
	}
}

func TestOutputBindByDest_RepeatedCalls(t *testing.T) {
	runSqlTest(t, testOutputBindByDestRepeatedCalls)
}

func testOutputBindByDestRepeatedCalls(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	t.tableName = "test_output_repeat"
	t.columnNameType = [][2]string{
		{"c1", "bigint"},
	}
	t.genTableTest()
	defer t.dropTable()

	for i := 0; i < 50; i++ {
		val := int64(i * 1000)
		var out int64
		t.mustExec(
			fmt.Sprintf(`insert into %s(c1) values(?) returning c1 into ?`, t.tableName),
			val,
			sql.Out{Dest: &out},
		)
		if out != val {
			t.Fatalf("iteration %d: got %d, want %d", i, out, val)
		}
	}
}

func TestOutputBindByDest_LargeStringOutput(t *testing.T) {
	runSqlTest(t, testOutputBindByDestLargeStringOutput)
}

func testOutputBindByDestLargeStringOutput(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	t.tableName = "test_output_large_str"
	t.columnNameType = [][2]string{
		{"c1", "varchar(32000)"},
	}
	t.genTableTest()
	defer t.dropTable()

	largeStr := strings.Repeat("x", 16000)
	var out string
	t.mustExec(
		fmt.Sprintf(`insert into %s(c1) values(?) returning c1 into ?`, t.tableName),
		largeStr,
		sql.Out{Dest: &out},
	)
	if len(out) != 16000 {
		t.Fatalf("large string: got len=%d, want 16000", len(out))
	}
	if out != largeStr {
		t.Fatal("large string content mismatch")
	}
}

func TestOutputBindByDest_NullString(t *testing.T) {
	runSqlTest(t, testOutputBindByDestNullString)
}

func testOutputBindByDestNullString(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	t.tableName = "test_output_null_str"
	t.columnNameType = [][2]string{
		{"c1", "varchar(100)"},
	}
	t.genTableTest()
	defer t.dropTable()

	var outStr string
	t.mustExec(
		fmt.Sprintf(`insert into %s(c1) values(?) returning c1 into ?`, t.tableName),
		nil,
		sql.Out{Dest: &outStr},
	)

	if outStr != "" {
		t.Fatalf("null string: got %q, want empty", outStr)
	}
}

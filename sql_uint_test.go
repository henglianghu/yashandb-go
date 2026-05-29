package yasdb

import (
	"fmt"
	"testing"
)

func TestUintInputParam(t *testing.T) {
	// t.Skip("only support YashanDB(MySQL)：go test -run TestUintInputParam --dsn 'sys/Cod-2022@127.0.0.1:1688?compat_vector=mysql' ")
	runSqlTest(t, testUintInputParam)
}

func testUintInputParam(t *sqlTest) {
	if t.IsYasdbMode() {
		t.Skip("only support YashanDB(MySQL)：go test -run TestUintInputParam --dsn 'sys/Cod-2022@127.0.0.1:1688?compat_vector=mysql' ")
	}
	si := sqlGenInfo{
		tableName: "test_uint_input_param",
		columnNameType: [][2]string{
			{"c1", "tinyint UNSIGNED"},
			{"c2", "smallint UNSIGNED"},
			{"c3", "INT UNSIGNED"},
			{"c4", "bigint UNSIGNED"},
		},
	}
	t.sqlGenInfo = &si

	t.genTableTest()
	defer t.dropTable()

	// 测试直接值插入
	t.mustExec(
		"insert into "+si.tableName+"(c1, c2, c3, c4) values(?,?,?,?)",
		uint8(255), uint16(65535), uint32(4294967295), 1,
	)

	// 测试变量插入
	var (
		v1 uint8  = 100
		v2 uint16 = 10000
		v3 uint32 = 1000000
		v4 uint64 = 10000000000
	)
	t.mustExec(
		"insert into "+si.tableName+"(c1, c2, c3, c4) values(?,?,?,?)",
		v1, v2, v3, v4,
	)

	// 测试命名参数
	t.mustExec(
		"insert into "+si.tableName+"(c1, c2, c3, c4) values(:1,:2,:3,:4)",
		uint8(1), uint16(2), uint32(3), uint64(4),
	)

	// 验证读取
	rows := t.mustQuery("select c1, c2, c3, c4 from " + si.tableName + " order by c1")
	defer rows.Close()
	ts, _ := rows.ColumnTypes()
	for _, t := range ts {
		fmt.Println(t.DatabaseTypeName())
	}

	count := 0
	for rows.Next() {
		var c1 uint8
		var c2 uint16
		var c3 uint32
		var c4 uint64
		if err := rows.Scan(&c1, &c2, &c3, &c4); err != nil {
			t.Fatalf("scan error: %v", err)
		}
		count++
	}
	if count != 3 {
		t.Fatalf("expected 3 rows, got %d", count)
	}
}

/*
func TestUintOutputParam(t *testing.T) {
	runSqlTest(t, testUintOutputParam)
}

func testUintOutputParam(t *sqlTest) {
	si := sqlGenInfo{
		tableName: "test_uint_output_param",
		columnNameType: [][2]string{
			{"c1", "tinyint UNSIGNED"},
			{"c2", "smallint UNSIGNED"},
			{"c3", "INT UNSIGNED"},
			{"c4", "bigint UNSIGNED"},
		},
	}
	t.sqlGenInfo = &si

	t.genTableTest()
	defer t.dropTable()

	// 先插入测试数据
	t.mustExec(
		"insert into "+si.tableName+"(c1, c2, c3, c4) values(?,?,?,?)",
		uint8(255), uint16(65535), uint32(4294967295), uint64(9223372036854775807),
	)

	// 测试 OUT 参数
	var outC1 uint8
	var outC2 uint16
	var outC3 uint32
	var outC4 uint64

	t.mustExec(
		`delete from `+si.tableName+` where c1 = ?
        returning c1, c2, c3, c4 into ?,?,?,?`,
		uint8(255),
		sql.Out{Dest: &outC1},
		sql.Out{Dest: &outC2},
		sql.Out{Dest: &outC3},
		sql.Out{Dest: &outC4},
	)

	if outC1 != 255 {
		t.Fatalf("expected outC1=255, got %d", outC1)
	}
	if outC2 != 65535 {
		t.Fatalf("expected outC2=65535, got %d", outC2)
	}
	if outC3 != 4294967295 {
		t.Fatalf("expected outC3=4294967295, got %d", outC3)
	}
	if outC4 != 9223372036854775807 {
		t.Fatalf("expected outC4=9223372036854775807, got %d", outC4)
	}
}

func TestUintMixedInputOutput(t *testing.T) {
	runSqlTest(t, testUintMixedInputOutput)
}

func testUintMixedInputOutput(t *sqlTest) {
	si := sqlGenInfo{
		tableName: "test_uint_mixed_io",
		columnNameType: [][2]string{
			{"c1", "tinyint UNSIGNED"},
			{"c2", "smallint UNSIGNED"},
			{"c3", "INT UNSIGNED"},
			{"c4", "bigint UNSIGNED"},
		},
	}
	t.sqlGenInfo = &si

	t.genTableTest()
	defer t.dropTable()

	// 插入初始数据
	t.mustExec(
		"insert into "+si.tableName+"(c1, c2, c3, c4) values(?,?,?,?)",
		uint8(10), uint16(200), uint32(3000), uint64(40000),
	)

	// 使用 INOUT 参数测试
	var inoutC1 uint8 = 100
	t.mustExec(
		`update `+si.tableName+` set c1 = ? where c1 = ?
        returning c1 into ?`,
		uint8(200),
		uint8(10),
		sql.Out{Dest: &inoutC1, In: true},
	)

	if inoutC1 != 200 {
		t.Fatalf("expected inoutC1=200, got %d", inoutC1)
	}
}
*/

func TestUintEdgeValues(t *testing.T) {
	runSqlTest(t, testUintEdgeValues)
}

func testUintEdgeValues(t *sqlTest) {
	if t.IsYasdbMode() {
		t.Skip("only support YashanDB(MySQL)：go test -run TestUintInputParam --dsn 'sys/Cod-2022@127.0.0.1:1688?compat_vector=mysql' ")
	}
	si := sqlGenInfo{
		tableName: "test_uint_edge_values",
		columnNameType: [][2]string{
			{"c1", "tinyint UNSIGNED"},
			{"c2", "smallint UNSIGNED"},
			{"c3", "INT UNSIGNED"},
			{"c4", "bigint UNSIGNED"},
		},
	}
	t.sqlGenInfo = &si

	t.genTableTest()
	defer t.dropTable()

	// 测试最小值
	t.mustExec(
		"insert into "+si.tableName+"(c1, c2, c3, c4) values(?,?,?,?)",
		uint8(0), uint16(0), uint32(0), uint64(0),
	)

	// 测试最大值
	t.mustExec(
		"insert into "+si.tableName+"(c1, c2, c3, c4) values(?,?,?,?)",
		uint8(255), uint16(65535), uint32(4294967295), uint64(9223372036854775807),
	)

	// 测试中间值
	t.mustExec(
		"insert into "+si.tableName+"(c1, c2, c3, c4) values(?,?,?,?)",
		uint8(127), uint16(32767), uint32(2147483647), uint64(9223372036854775807),
	)

	// 验证
	rows := t.mustQuery("select c1, c2, c3, c4 from " + si.tableName + " order by c1")
	defer rows.Close()

	expected := [][4]interface{}{
		{uint8(0), uint16(0), uint32(0), uint64(0)},
		{uint8(127), uint16(32767), uint32(2147483647), uint64(9223372036854775807)},
		{uint8(255), uint16(65535), uint32(4294967295), uint64(9223372036854775807)},
	}

	rowIdx := 0
	for rows.Next() {
		var c1 uint8
		var c2 uint16
		var c3 uint32
		var c4 uint64
		if err := rows.Scan(&c1, &c2, &c3, &c4); err != nil {
			t.Fatalf("scan error: %v", err)
		}
		if rowIdx >= len(expected) {
			t.Fatalf("unexpected row %d", rowIdx)
		}
		exp := expected[rowIdx]
		if c1 != exp[0] || c2 != exp[1] || c3 != exp[2] || c4 != exp[3] {
			t.Fatalf("row %d mismatch: got (%d,%d,%d,%d), expected (%v,%v,%v,%v)",
				rowIdx, c1, c2, c3, c4, exp[0], exp[1], exp[2], exp[3])
		}
		rowIdx++
	}
	if rowIdx != len(expected) {
		t.Fatalf("expected %d rows, got %d", len(expected), rowIdx)
	}
}

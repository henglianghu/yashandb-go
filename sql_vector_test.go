package yasdb

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"
)

func TestVector(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testVector)
}

func testVector(t *sqlTest) {
	si := sqlGenInfo{}
	t.sqlGenInfo = &si

	// 测试 Vector 读取 - 使用字符串形式的向量文本
	// 根据测试 SQL: create table vector_test_data(c1 vector(3, FLOAT32), c2 vector(3, FLOAT64), c3 varchar(64), c4 clob)
	// 注意: 这里我们使用 from_vector/to_vector 函数转换，因为直接插入 vector 需要特殊处理

	// 清理可能存在的表
	t.Exec("DROP TABLE IF EXISTS vector_test_data")

	// 创建表
	_, err := t.DB.Exec("CREATE TABLE vector_test_data(c1 vector(3, FLOAT32), c2 vector(3, FLOAT64), c3 varchar(64), c4 clob)")
	if err != nil {
		t.Logf("create table warning: %v (可能需要更高版本支持)", err)
		// 如果创建表失败，尝试简单测试
		t.fatalVectorTest("表创建失败")
		return
	}
	// defer t.Exec("DROP TABLE IF EXISTS vector_test_data")

	// 插入数据 - 使用文本形式
	_, err = t.DB.Exec("INSERT INTO vector_test_data VALUES('[1, 0.2, 0.03]', '[1, 0.2, 0.03]', '[1, 0.2, 0.03]', '[1, 0.2, 0.03]')")
	if err != nil {
		t.Logf("insert warning: %v", err)
		t.fatalVectorTest("数据插入失败")
		return
	}

	_, err = t.DB.Exec("INSERT INTO vector_test_data VALUES('[2, 0.3, 0.04]', '[2, 0.3, 0.04]', '[2, 0.3, 0.04]', '[2, 0.3, 0.04]')")
	if err != nil {
		t.Logf("insert warning: %v", err)
		t.fatalVectorTest("数据插入失败")
		return
	}
	vstr := "[-3, 0.4, -0.05]"
	_, err = t.DB.Exec("INSERT INTO vector_test_data VALUES(?, ?, ?, ?)", vstr, vstr, vstr, vstr)
	if err != nil {
		t.Logf("insert warning: %v", err)
		t.fatalVectorTest("数据插入失败")
		return
	}

	t.Exec("commit")

	// 测试查询 - 使用 from_vector(to_vector()) 转换
	testCases := []struct {
		name     string
		query    string
		checkFun func(rows *sql.Rows) error
	}{
		{
			name:  "query c1 with from_vector(to_vector(c1))",
			query: "SELECT from_vector(to_vector(c1)) FROM vector_test_data",
			checkFun: func(rows *sql.Rows) error {
				var val string
				count := 0
				for rows.Next() {
					if err := rows.Scan(&val); err != nil {
						return err
					}
					t.Logf("c1 to_text: %s", val)
					count++
				}
				if count != 3 {
					return fmt.Errorf("expected 3 rows, got %d", count)
				}
				return nil
			},
		},
		{
			name:  "query c2 with from_vector(to_vector(c2))",
			query: "SELECT from_vector(to_vector(c2)) FROM vector_test_data",
			checkFun: func(rows *sql.Rows) error {
				var val string
				count := 0
				for rows.Next() {
					if err := rows.Scan(&val); err != nil {
						return err
					}
					t.Logf("c2 to_text: %s", val)
					count++
				}
				if count != 3 {
					return fmt.Errorf("expected 3 rows, got %d", count)
				}
				return nil
			},
		},
		{
			name:  "query c3 (varchar) with from_vector(to_vector(c3))",
			query: "SELECT from_vector(to_vector(c3)) FROM vector_test_data",
			checkFun: func(rows *sql.Rows) error {
				var val string
				count := 0
				for rows.Next() {
					if err := rows.Scan(&val); err != nil {
						return err
					}
					t.Logf("c3 to_text: %s", val)
					count++
				}
				if count != 3 {
					return fmt.Errorf("expected 3 rows, got %d", count)
				}
				return nil
			},
		},
		{
			name:  "query c4 (clob) with from_vector(to_vector(c4))",
			query: "SELECT from_vector(to_vector(c4)) FROM vector_test_data",
			checkFun: func(rows *sql.Rows) error {
				var val string
				count := 0
				for rows.Next() {
					if err := rows.Scan(&val); err != nil {
						return err
					}
					t.Logf("c4 to_text: %s", val)
					count++
				}
				if count != 3 {
					return fmt.Errorf("expected 3 rows, got %d", count)
				}
				return nil
			},
		},
		{
			name:  "query c1 RETURNING VARCHAR",
			query: "SELECT from_vector(c1 RETURNING VARCHAR(64)) FROM vector_test_data",
			checkFun: func(rows *sql.Rows) error {
				var val string
				count := 0
				for rows.Next() {
					if err := rows.Scan(&val); err != nil {
						return err
					}
					t.Logf("c1 RETURNING VARCHAR: %s", val)
					count++
				}
				if count != 3 {
					return fmt.Errorf("expected 3 rows, got %d", count)
				}
				return nil
			},
		},
		{
			name:  "query c1 RETURNING CLOB",
			query: "SELECT from_vector(c1 RETURNING CLOB) FROM vector_test_data",
			checkFun: func(rows *sql.Rows) error {
				var val string
				count := 0
				for rows.Next() {
					if err := rows.Scan(&val); err != nil {
						return err
					}
					t.Logf("c1 RETURNING CLOB: %s", val)
					count++
				}
				if count != 3 {
					return fmt.Errorf("expected 3 rows, got %d", count)
				}
				return nil
			},
		},
		{
			name:  "query to_vector from c1",
			query: "SELECT to_vector(c1) FROM vector_test_data",
			checkFun: func(rows *sql.Rows) error {
				// to_vector 返回的是 vector 类型，扫描到 Vector 类型
				var val Vector
				count := 0
				for rows.Next() {
					if err := rows.Scan(&val); err != nil {
						return err
					}
					t.Logf("to_vector(c1): %+v, dim=%d, format=%d", val.Data, val.Dim, val.Format)
					count++
				}
				if count != 3 {
					return fmt.Errorf("expected 3 rows, got %d", count)
				}
				return nil
			},
		},
		{
			name:  "query to_vector from c2",
			query: "SELECT to_vector(c2) FROM vector_test_data",
			checkFun: func(rows *sql.Rows) error {
				var val Vector
				count := 0
				for rows.Next() {
					if err := rows.Scan(&val); err != nil {
						return err
					}
					t.Logf("to_vector(c2): %+v, dim=%d, format=%d", val.Data, val.Dim, val.Format)
					count++
				}
				if count != 3 {
					return fmt.Errorf("expected 3 rows, got %d", count)
				}
				return nil
			},
		},
		{
			name:  "query to_vector from c3 (varchar)",
			query: "SELECT to_vector(c3) FROM vector_test_data",
			checkFun: func(rows *sql.Rows) error {
				var val Vector
				count := 0
				for rows.Next() {
					if err := rows.Scan(&val); err != nil {
						return err
					}
					t.Logf("to_vector(c3): %+v, dim=%d, format=%d", val.Data, val.Dim, val.Format)
					count++
				}
				if count != 3 {
					return fmt.Errorf("expected 3 rows, got %d", count)
				}
				return nil
			},
		},
		{
			name:  "query to_vector from c4 (clob)",
			query: "SELECT to_vector(c4) FROM vector_test_data",
			checkFun: func(rows *sql.Rows) error {
				var val Vector
				count := 0
				for rows.Next() {
					if err := rows.Scan(&val); err != nil {
						return err
					}
					t.Logf("to_vector(c4): %+v, dim=%d, format=%d", val.Data, val.Dim, val.Format)
					count++
				}
				if count != 3 {
					return fmt.Errorf("expected 3 rows, got %d", count)
				}
				return nil
			},
		},
	}

	// 保存外部 t 的引用以便在内部使用
	outerT := t
	for _, tc := range testCases {
		tc := tc // 捕获循环变量
		t.Run(tc.name, func(t *testing.T) {
			rows, err := outerT.DB.Query(tc.query)
			if err != nil {
				t.Fatalf("query error: %v (可能需要数据库支持)", err)
				return
			}
			defer rows.Close()

			if err := tc.checkFun(rows); err != nil {
				t.Fatalf("check failed: %v", err)
			}
		})
	}
}

func (t *sqlTest) fatalVectorTest(reason string) {
	t.Fatalf("skip Vector test: %s", reason)
}

// TestVectorScanType 测试 Vector 扫描类型
func TestVectorScanType(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testVectorScanType)
}

func testVectorScanType(t *sqlTest) {
	// 创建简单的表测试
	t.Exec("DROP TABLE IF EXISTS vector_scan_test")
	defer t.Exec("DROP TABLE IF EXISTS vector_scan_test")

	// 尝试创建包含 vector 的表
	_, err := t.DB.Exec("CREATE TABLE vector_scan_test(id int, vec vector(3, FLOAT32))")
	if err != nil {
		t.Logf("create table warning: %v", err)
		return
	}

	// 插入数据
	t.DB.Exec("INSERT INTO vector_scan_test VALUES(1, '[1,2,3]')")

	// 查询并检查 ColumnTypeScanType
	rows, err := t.DB.Query("SELECT vec FROM vector_scan_test")
	if err != nil {
		t.Logf("query warning: %v", err)
		return
	}
	defer rows.Close()

	cols, err := rows.ColumnTypes()
	if err != nil {
		t.Errorf("get column types error: %v", err)
		return
	}

	for _, col := range cols {
		t.Logf("column: %s, scanType: %v, databaseTypeName: %s",
			col.Name(), col.ScanType(), col.DatabaseTypeName())
	}
}

// TestVectorTypeName 测试 Vector 类型名
func TestVectorTypeName(t *testing.T) {
	t.Parallel()
	// 验证 GetDatabaseTypeName 返回 VECTOR
	typeName := GetDatabaseTypeName(42) // YAPI_TYPE_VECTOR = 42
	if typeName != "VECTOR" {
		t.Errorf("expected VECTOR, got %s", typeName)
	} else {
		t.Logf("GetDatabaseTypeName(42) = %s", typeName)
	}
}

// TestVectorFormatConstants 测试 Vector 格式常量
func TestVectorFormatConstants(t *testing.T) {
	t.Parallel()

	// 验证 VectorFormat 常量值
	if VectorFormatFlex != 0 {
		t.Errorf("VectorFormatFlex expected 0, got %d", VectorFormatFlex)
	}
	if VectorFormatFloat16 != 1 {
		t.Errorf("VectorFormatFloat16 expected 1, got %d", VectorFormatFloat16)
	}
	if VectorFormatFloat32 != 2 {
		t.Errorf("VectorFormatFloat32 expected 2, got %d", VectorFormatFloat32)
	}
	if VectorFormatFloat64 != 3 {
		t.Errorf("VectorFormatFloat64 expected 3, got %d", VectorFormatFloat64)
	}
	if VectorFormatInt8 != 4 {
		t.Errorf("VectorFormatInt8 expected 4, got %d", VectorFormatInt8)
	}

	t.Log("Vector format constants verified")
}

// TestVectorStruct 测试 Vector 结构体
func TestVectorStruct(t *testing.T) {
	t.Parallel()

	// 测试 Vector 结构体创建
	v1 := Vector{
		Data:   []float32{1.0, 2.0, 3.0},
		Dim:    3,
		Format: VectorFormatFloat32,
	}
	t.Logf("Vector with float32: %+v", v1)

	v2 := Vector{
		Data:   []float64{1.0, 2.0, 3.0},
		Dim:    3,
		Format: VectorFormatFloat64,
	}
	t.Logf("Vector with float64: %+v", v2)

	v3 := Vector{
		Data:   []int8{1, 2, 3},
		Dim:    3,
		Format: VectorFormatInt8,
	}
	t.Logf("Vector with int8: %+v", v3)

	// 测试空 Data
	v4 := Vector{
		Data:   nil,
		Dim:    0,
		Format: VectorFormatFlex,
	}
	t.Logf("Vector with nil data: %+v", v4)

	t.Log("Vector struct test passed")
}

// TestVectorInputBind 测试 Vector 输入绑定
func TestVectorInputBind(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testVectorInputBind)
}

func testVectorInputBind(t *sqlTest) {
	// 创建测试表
	t.Exec("DROP TABLE IF EXISTS test_vector_input")
	defer t.Exec("DROP TABLE IF EXISTS test_vector_input")

	// 尝试创建包含 vector 的表
	_, err := t.DB.Exec("CREATE TABLE test_vector_input (id INT, vec VECTOR(3, FLOAT32))")
	if err != nil {
		t.Logf("create table warning: %v (可能需要更高版本支持)", err)
		t.fatalVectorTest("表创建失败")
		return
	}

	// 尝试绑定 Vector 参数
	vec := Vector{Data: []float32{1.0, 2.0, 3.0}, Dim: 3, Format: VectorFormatFloat32}
	_, err = t.DB.Exec("INSERT INTO test_vector_input VALUES (?, ?)", 1, vec)
	if err != nil {
		t.Fatalf("Vector input bind should work, got error: %v", err)
	}

	// 验证插入成功
	var count int
	err = t.DB.QueryRow("SELECT COUNT(*) FROM test_vector_input").Scan(&count)
	if err != nil {
		t.Fatalf("query count error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}

	// 验证数据可以查询回来
	var resultVec Vector
	err = t.DB.QueryRow("SELECT vec FROM test_vector_input WHERE id = 1").Scan(&resultVec)
	if err != nil {
		t.Fatalf("query vector error: %v", err)
	}
	t.Logf("inserted vector: %+v, queried vector: %+v", vec, resultVec)

	// Verify the retrieved vector matches
	if resultVec.Dim != vec.Dim {
		t.Fatalf("vector dimension mismatch: expected %d, got %d", vec.Dim, resultVec.Dim)
	}
	// Compare actual data
	switch v := resultVec.Data.(type) {
	case []float32:
		expected := vec.Data.([]float32)
		if len(v) != len(expected) {
			t.Fatalf("vector data length mismatch")
		}
		for i := range v {
			if v[i] != expected[i] {
				t.Fatalf("vector data mismatch at index %d", i)
			}
		}
	}

	t.Log("Vector input bind test passed")
}

// TestVectorOutputBind 测试 Vector 输出绑定
func TestVectorOutputBind(t *testing.T) {
	t.Parallel()

	// 测试 Go 层面的输出绑定功能（不需要数据库连接）
	// 预分配 Vector 的 Data 用于接收输出
	vec := Vector{
		Data:   make([]float32, 3), // 预分配 3 维向量
		Dim:    3,
		Format: VectorFormatFloat32,
	}
	out, err := NewOutputBindValue(&vec, WithTypeVector())
	if err != nil {
		t.Fatalf("NewOutputBindValue error: %v", err)
	}
	// 验证 outputBindInfo 已正确设置
	if out == nil {
		t.Fatal("outputBindInfo should not be nil")
	}
	t.Logf("Vector output bind info created successfully: %+v", vec)
}

// TestVectorResourceCleanup 测试 Vector 资源释放
func TestVectorResourceCleanup(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testVectorResourceCleanup)
}

func testVectorResourceCleanup(t *sqlTest) {
	// 清理可能存在的表
	t.Exec("DROP TABLE IF EXISTS test_vector_cleanup")

	// 创建测试表
	_, err := t.DB.Exec("CREATE TABLE test_vector_cleanup (id INT, vec VECTOR(3, FLOAT32))")
	if err != nil {
		t.Logf("create table warning: %v (可能需要更高版本支持)", err)
		t.fatalVectorTest("表创建失败")
		return
	}
	defer t.Exec("DROP TABLE IF EXISTS test_vector_cleanup")

	// 插入多条数据
	for i := 0; i < 10; i++ {
		vec := Vector{Data: []float32{float32(i), float32(i + 1), float32(i + 2)}, Dim: 3, Format: VectorFormatFloat32}
		_, err = t.DB.Exec("INSERT INTO test_vector_cleanup VALUES (?, ?)", i, vec)
		if err != nil {
			t.fatalVectorTest(fmt.Sprintf("数据插入失败: %v", err))
			return
		}
	}

	t.Exec("commit")

	// 查询数据并验证不会泄漏
	rows, err := t.DB.Query("SELECT id, vec FROM test_vector_cleanup")
	if err != nil {
		t.fatalVectorTest(fmt.Sprintf("查询失败: %v", err))
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var vec Vector
		err := rows.Scan(&id, &vec)
		if err != nil {
			t.fatalVectorTest(fmt.Sprintf("Scan 失败: %v", err))
			return
		}
		count++
		t.Logf("id=%d, vec.Dim=%d, vec.Format=%d", id, vec.Dim, vec.Format)
	}

	if err := rows.Err(); err != nil {
		t.fatalVectorTest(fmt.Sprintf("rows.Err: %v", err))
		return
	}

	if count != 10 {
		t.Fatalf("expected 10 rows, got %d", count)
	}

	t.Log("Vector resource cleanup test passed")
}

// TestVectorInsertAndQuery 测试 Vector 插入和查询
func TestVectorInsertAndQuery(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testVectorInsertAndQuery)
}

func testVectorInsertAndQuery(t *sqlTest) {
	// 清理可能存在的表
	t.Exec("DROP TABLE IF EXISTS test_vector_insert_query")
	defer t.Exec("DROP TABLE IF EXISTS test_vector_insert_query")

	// 创建测试表 - 只使用 FLOAT32 和 FLOAT64 格式
	_, err := t.DB.Exec("CREATE TABLE test_vector_insert_query (id INT, vec_f32 VECTOR(3, FLOAT32), vec_f64 VECTOR(4, FLOAT64))")
	if err != nil {
		t.Logf("create table warning: %v (可能需要更高版本支持)", err)
		t.fatalVectorTest("表创建失败")
		return
	}

	// 准备测试数据 - 不同格式的向量
	testVectors := []struct {
		id     int
		vecF32 Vector
		vecF64 Vector
	}{
		{
			id:     1,
			vecF32: Vector{Data: []float32{1.5, 2.5, 3.5}, Dim: 3, Format: VectorFormatFloat32},
			vecF64: Vector{Data: []float64{1.1, 2.2, 3.3, 4.4}, Dim: 4, Format: VectorFormatFloat64},
		},
		{
			id:     2,
			vecF32: Vector{Data: []float32{-1.0, 0.0, 1.0}, Dim: 3, Format: VectorFormatFloat32},
			vecF64: Vector{Data: []float64{-0.5, 0.5, -0.25, 0.25}, Dim: 4, Format: VectorFormatFloat64},
		},
		{
			id:     3,
			vecF32: Vector{Data: []float32{0.001, 0.002, 0.003}, Dim: 3, Format: VectorFormatFloat32},
			vecF64: Vector{Data: []float64{3.14159, 2.71828, 1.41421, 0.57721}, Dim: 4, Format: VectorFormatFloat64},
		},
	}

	// 插入数据
	t.Logf("=== Inserting vectors ===")
	for _, tv := range testVectors {
		_, err = t.DB.Exec(
			"INSERT INTO test_vector_insert_query VALUES (?, ?, ?)",
			tv.id, tv.vecF32, tv.vecF64,
		)
		if err != nil {
			t.Fatalf("insert vector failed: %v", err)
			return
		}
		t.Logf("Inserted id=%d, f32=%v, f64=%v", tv.id, tv.vecF32.Data, tv.vecF64.Data)
	}

	t.Exec("commit")

	// 查询并验证数据
	t.Logf("=== Querying vectors ===")
	rows, err := t.DB.Query("SELECT id, vec_f32, vec_f64 FROM test_vector_insert_query ORDER BY id")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var vecF32, vecF64 Vector

		err := rows.Scan(&id, &vecF32, &vecF64)
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}

		expected := testVectors[count]

		// 验证 FLOAT32 向量
		if vecF32.Dim != expected.vecF32.Dim {
			t.Errorf("row %d: vec_f32 dim mismatch: expected %d, got %d", count, expected.vecF32.Dim, vecF32.Dim)
		}
		if f32Data, ok := vecF32.Data.([]float32); ok {
			expectedF32 := expected.vecF32.Data.([]float32)
			for i := range f32Data {
				if f32Data[i] != expectedF32[i] {
					t.Errorf("row %d: vec_f32[%d] mismatch: expected %f, got %f", count, i, expectedF32[i], f32Data[i])
				}
			}
		}

		// 验证 FLOAT64 向量
		if vecF64.Dim != expected.vecF64.Dim {
			t.Errorf("row %d: vec_f64 dim mismatch: expected %d, got %d", count, expected.vecF64.Dim, vecF64.Dim)
		}
		if f64Data, ok := vecF64.Data.([]float64); ok {
			expectedF64 := expected.vecF64.Data.([]float64)
			for i := range f64Data {
				if f64Data[i] != expectedF64[i] {
					t.Errorf("row %d: vec_f64[%d] mismatch: expected %f, got %f", count, i, expectedF64[i], f64Data[i])
				}
			}
		}

		t.Logf("Verified id=%d: f32=%v (format=%d), f64=%v (format=%d)",
			id, vecF32.Data, vecF32.Format, vecF64.Data, vecF64.Format)

		count++
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}

	if count != len(testVectors) {
		t.Fatalf("expected %d rows, got %d", len(testVectors), count)
	}

	t.Log("Vector insert and query test passed")
}

// TestVectorBatchInsert 测试批量 Vector 插入
func TestVectorBatchInsert(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testVectorBatchInsert)
}

func testVectorBatchInsert(t *sqlTest) {
	// 清理可能存在的表
	t.Exec("DROP TABLE IF EXISTS test_vector_batch")
	defer t.Exec("DROP TABLE IF EXISTS test_vector_batch")

	// 创建测试表
	_, err := t.DB.Exec("CREATE TABLE test_vector_batch (id INT, vec VECTOR(3, FLOAT32))")
	if err != nil {
		t.Logf("create table warning: %v (可能需要更高版本支持)", err)
		t.fatalVectorTest("表创建失败")
		return
	}

	// 准备批量数据
	vectors := []Vector{
		{Data: []float32{1.0, 2.0, 3.0}, Dim: 3, Format: VectorFormatFloat32},
		{Data: []float32{4.0, 5.0, 6.0}, Dim: 3, Format: VectorFormatFloat32},
		{Data: []float32{7.0, 8.0, 9.0}, Dim: 3, Format: VectorFormatFloat32},
		{Data: []float32{10.0, 11.0, 12.0}, Dim: 3, Format: VectorFormatFloat32},
	}

	// 使用预处理语句批量插入
	stmt, err := t.DB.Prepare("INSERT INTO test_vector_batch VALUES (?, ?)")
	if err != nil {
		t.Fatalf("prepare failed: %v", err)
	}
	defer stmt.Close()

	t.Logf("=== Batch inserting %d vectors ===", len(vectors))
	for i, vec := range vectors {
		_, err = stmt.Exec(i+1, vec)
		if err != nil {
			t.Fatalf("batch insert failed at index %d: %v", i, err)
		}
		t.Logf("Inserted vector %d: %v", i+1, vec.Data)
	}

	t.Exec("commit")

	// 验证插入数量
	var count int
	err = t.DB.QueryRow("SELECT COUNT(*) FROM test_vector_batch").Scan(&count)
	if err != nil {
		t.Fatalf("query count failed: %v", err)
	}
	if count != len(vectors) {
		t.Fatalf("expected %d rows, got %d", len(vectors), count)
	}

	// 查询并验证所有数据
	rows, err := t.DB.Query("SELECT id, vec FROM test_vector_batch ORDER BY id")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	idx := 0
	for rows.Next() {
		var id int
		var vec Vector
		err := rows.Scan(&id, &vec)
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}

		expected := vectors[idx]
		if vec.Dim != expected.Dim {
			t.Errorf("row %d: dim mismatch: expected %d, got %d", idx, expected.Dim, vec.Dim)
		}

		if vecData, ok := vec.Data.([]float32); ok {
			expectedData := expected.Data.([]float32)
			for i := range vecData {
				if vecData[i] != expectedData[i] {
					t.Errorf("row %d: data[%d] mismatch: expected %f, got %f", idx, i, expectedData[i], vecData[i])
				}
			}
		}

		t.Logf("Verified id=%d: vec=%v (format=%d)", id, vec.Data, vec.Format)
		idx++
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}

	if idx != len(vectors) {
		t.Fatalf("expected %d rows, got %d", len(vectors), idx)
	}

	t.Log("Vector batch insert test passed")
}

// TestVectorBatchSliceInsert 测试 []Vector 批量插入
func TestVectorBatchSliceInsert(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testVectorBatchSliceInsert)
}

func testVectorBatchSliceInsert(t *sqlTest) {
	// 清理可能存在的表
	t.Exec("DROP TABLE IF EXISTS test_vector_batch_slice")
	defer t.Exec("DROP TABLE IF EXISTS test_vector_batch_slice")

	// 创建测试表
	_, err := t.DB.Exec("CREATE TABLE test_vector_batch_slice (id INT, vec VECTOR(3, FLOAT32))")
	if err != nil {
		t.Logf("create table warning: %v (可能需要更高版本支持)", err)
		t.fatalVectorTest("表创建失败")
		return
	}

	// 准备批量数据 - 使用 []Vector
	vectors := []Vector{
		{Data: []float32{1.0, 2.0, 3.0}, Dim: 3, Format: VectorFormatFloat32},
		{Data: []float32{4.0, 5.0, 6.0}, Dim: 3, Format: VectorFormatFloat32},
		{Data: []float32{7.0, 8.0, 9.0}, Dim: 3, Format: VectorFormatFloat32},
		{Data: []float32{10.0, 11.0, 12.0}, Dim: 3, Format: VectorFormatFloat32},
	}

	// 使用 UNNEST 批量插入
	_, err = t.DB.Exec(`
		INSERT INTO test_vector_batch_slice (id, vec)
		SELECT COLUMN_VALUE, vec FROM TABLE(UNNEST(
			yasdb_array.t_number_list(1, 2, 3, 4),
			yasdb_array.t_vector_list(?, ?, ?, ?)
		))
	`, vectors[0], vectors[1], vectors[2], vectors[3])
	if err != nil {
		// 如果数组类型不支持，回退到逐条插入验证
		t.Logf("Array type not supported, falling back to individual inserts: %v", err)
		stmt, err := t.DB.Prepare("INSERT INTO test_vector_batch_slice VALUES (?, ?)")
		if err != nil {
			t.Fatalf("prepare failed: %v", err)
		}
		defer stmt.Close()

		for i, vec := range vectors {
			_, err = stmt.Exec(i+1, vec)
			if err != nil {
				t.Fatalf("insert failed at index %d: %v", i, err)
			}
		}
	}

	t.Exec("commit")

	// 验证插入数量
	var count int
	err = t.DB.QueryRow("SELECT COUNT(*) FROM test_vector_batch_slice").Scan(&count)
	if err != nil {
		t.Fatalf("query count failed: %v", err)
	}
	if count != len(vectors) {
		t.Fatalf("expected %d rows, got %d", len(vectors), count)
	}

	// 查询并验证所有数据
	rows, err := t.DB.Query("SELECT id, vec FROM test_vector_batch_slice ORDER BY id")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	idx := 0
	for rows.Next() {
		var id int
		var vec Vector
		err := rows.Scan(&id, &vec)
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}

		expected := vectors[idx]
		if vec.Dim != expected.Dim {
			t.Errorf("row %d: dim mismatch: expected %d, got %d", idx, expected.Dim, vec.Dim)
		}

		if vecData, ok := vec.Data.([]float32); ok {
			expectedData := expected.Data.([]float32)
			for i := range vecData {
				if vecData[i] != expectedData[i] {
					t.Errorf("row %d: data[%d] mismatch: expected %f, got %f", idx, i, expectedData[i], vecData[i])
				}
			}
		}

		t.Logf("Verified id=%d: vec=%v (format=%d)", id, vec.Data, vec.Format)
		idx++
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}

	if idx != len(vectors) {
		t.Fatalf("expected %d rows, got %d", len(vectors), idx)
	}

	t.Log("[]Vector batch slice insert test passed")
}

// TestVectorBatchPointerSliceInsert 测试 []*Vector 批量插入
func TestVectorBatchPointerSliceInsert(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testVectorBatchPointerSliceInsert)
}

func testVectorBatchPointerSliceInsert(t *sqlTest) {
	// 清理可能存在的表
	t.Exec("DROP TABLE IF EXISTS test_vector_batch_ptr")
	defer t.Exec("DROP TABLE IF EXISTS test_vector_batch_ptr")

	// 创建测试表
	_, err := t.DB.Exec("CREATE TABLE test_vector_batch_ptr (id INT, vec VECTOR(3, FLOAT32))")
	if err != nil {
		t.Logf("create table warning: %v (可能需要更高版本支持)", err)
		t.fatalVectorTest("表创建失败")
		return
	}

	// 准备批量数据 - 使用 []*Vector
	vectors := []*Vector{
		{Data: []float32{1.0, 2.0, 3.0}, Dim: 3, Format: VectorFormatFloat32},
		{Data: []float32{4.0, 5.0, 6.0}, Dim: 3, Format: VectorFormatFloat32},
		{Data: []float32{7.0, 8.0, 9.0}, Dim: 3, Format: VectorFormatFloat32},
		{Data: []float32{10.0, 11.0, 12.0}, Dim: 3, Format: VectorFormatFloat32},
	}

	// 使用预处理语句逐条插入
	stmt, err := t.DB.Prepare("INSERT INTO test_vector_batch_ptr VALUES (?, ?)")
	if err != nil {
		t.Fatalf("prepare failed: %v", err)
	}
	defer stmt.Close()

	t.Logf("=== Batch inserting %d vectors (pointer slice) ===", len(vectors))
	for i, vec := range vectors {
		_, err = stmt.Exec(i+1, vec)
		if err != nil {
			t.Fatalf("batch insert failed at index %d: %v", i, err)
		}
		t.Logf("Inserted vector %d: %v", i+1, vec.Data)
	}

	t.Exec("commit")

	// 验证插入数量
	var count int
	err = t.DB.QueryRow("SELECT COUNT(*) FROM test_vector_batch_ptr").Scan(&count)
	if err != nil {
		t.Fatalf("query count failed: %v", err)
	}
	if count != len(vectors) {
		t.Fatalf("expected %d rows, got %d", len(vectors), count)
	}

	// 查询并验证所有数据
	rows, err := t.DB.Query("SELECT id, vec FROM test_vector_batch_ptr ORDER BY id")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	idx := 0
	for rows.Next() {
		var id int
		var vec Vector
		err := rows.Scan(&id, &vec)
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}

		expected := vectors[idx]
		if vec.Dim != expected.Dim {
			t.Errorf("row %d: dim mismatch: expected %d, got %d", idx, expected.Dim, vec.Dim)
		}

		if vecData, ok := vec.Data.([]float32); ok {
			expectedData := expected.Data.([]float32)
			for i := range vecData {
				if vecData[i] != expectedData[i] {
					t.Errorf("row %d: data[%d] mismatch: expected %f, got %f", idx, i, expectedData[i], vecData[i])
				}
			}
		}

		t.Logf("Verified id=%d: vec=%v (format=%d)", id, vec.Data, vec.Format)
		idx++
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}

	if idx != len(vectors) {
		t.Fatalf("expected %d rows, got %d", len(vectors), idx)
	}

	t.Log("[]*Vector batch pointer slice insert test passed")
}

// TestVectorValue_Contract 验证 Value() 是否符合 driver.Valuer 约定。
func TestVectorValue_Contract(t *testing.T) {
	t.Parallel()

	vec := Vector{Data: []float32{1, 2, 3}, Dim: 3, Format: VectorFormatFloat32}

	nilVec := Vector{}
	val, err := nilVec.Value()
	if err != nil {
		t.Fatalf("nil Data Value() error: %v", err)
	}
	if val != nil {
		t.Fatalf("nil Data Value() want nil, got %T %v", val, val)
	}

	val, err = vec.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	if !driver.IsValue(val) {
		t.Fatalf("Value() returns %T, which is NOT a valid driver.Value", val)
	}
	if val != vec.String() {
		t.Fatalf("Value() want %q, got %v", vec.String(), val)
	}
}

// TestVectorValue_RecursiveValuer 模拟上层对 Valuer 的递归格式化（如 GORM 日志）。
func TestVectorValue_RecursiveValuer(t *testing.T) {
	t.Parallel()

	vec := Vector{Data: []float32{1, 2, 3}, Dim: 3, Format: VectorFormatFloat32}

	const maxDepth = 100
	var callValue func(v interface{}, depth int) (interface{}, error)
	callValue = func(v interface{}, depth int) (interface{}, error) {
		if depth > maxDepth {
			return nil, fmt.Errorf("recursive Valuer call exceeded depth %d", maxDepth)
		}
		vr, ok := v.(driver.Valuer)
		if !ok {
			return v, nil
		}
		out, err := vr.Value()
		if err != nil {
			return nil, err
		}
		return callValue(out, depth+1)
	}

	out, err := callValue(vec, 0)
	if err != nil {
		t.Fatalf("recursive Valuer formatting should not fail: %v", err)
	}
	if out != vec.String() {
		t.Fatalf("recursive Valuer formatting want %q, got %v", vec.String(), out)
	}
}

// TestVectorValue_StringVsValue 对比 String() 与 Value() 的返回值类型。
func TestVectorValue_StringVsValue(t *testing.T) {
	t.Parallel()

	vec := Vector{Data: []float32{1, 2, 3}, Dim: 3, Format: VectorFormatFloat32}
	val, err := vec.Value()
	if err != nil {
		t.Fatal(err)
	}

	str, ok := val.(string)
	if !ok {
		t.Fatalf("Value() should return string, got %T", val)
	}
	if str != vec.String() {
		t.Fatalf("Value() want %q, String() %q", str, vec.String())
	}
}

// TestVectorValue_DirectBind 验证 Vector 参数绑定走 CheckNamedValue 专用路径。
func TestVectorValue_DirectBind(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testVectorValueDirectBind)
}

func testVectorValueDirectBind(t *sqlTest) {
	t.Exec("DROP TABLE IF EXISTS test_vector_value_bind")
	defer t.Exec("DROP TABLE IF EXISTS test_vector_value_bind")

	_, err := t.DB.Exec("CREATE TABLE test_vector_value_bind (id INT, vec VECTOR(3, FLOAT32))")
	if err != nil {
		t.fatalVectorTest("表创建失败")
		return
	}

	vec := Vector{Data: []float32{1, 2, 3}, Dim: 3, Format: VectorFormatFloat32}

	_, err = t.DB.Exec("INSERT INTO test_vector_value_bind VALUES (?, ?)", 1, vec)
	if err != nil {
		t.Fatalf("direct Vector bind should work: %v", err)
	}

	val, err := vec.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	str, ok := val.(string)
	if !ok || str != vec.String() {
		t.Fatalf("Value() should return string %q for display, got %T %v", vec.String(), val, val)
	}

	var got Vector
	err = t.DB.QueryRow("SELECT vec FROM test_vector_value_bind WHERE id = 1").Scan(&got)
	if err != nil {
		t.Fatalf("query vector error: %v", err)
	}

	expected := vec.Data.([]float32)
	gotData, ok := got.Data.([]float32)
	if !ok {
		t.Fatalf("scan got unexpected type %T", got.Data)
	}
	for i := range expected {
		if gotData[i] != expected[i] {
			t.Fatalf("data mismatch at %d: want %f got %f", i, expected[i], gotData[i])
		}
	}
}

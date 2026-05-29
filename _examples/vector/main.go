package main

import (
	"database/sql"
	"flag"
	"fmt"

	yasdb "github.com/yashan-technologies/yashandb-go"
)

func main() {
	dsn := flag.String("dsn", "", "dsn")
	flag.Parse()
	if *dsn == "" {
		fmt.Println("Usage: go run main.go -dsn 'user/password@host:port'")
		return
	}

	db, err := sql.Open("yasdb", *dsn)
	if err != nil {
		fmt.Println("failed to connect:", err)
		return
	}
	defer db.Close()

	// 清理可能存在的表
	_, err = db.Exec("DROP TABLE IF EXISTS vector_example")
	if err != nil {
		fmt.Println("drop table error:", err)
	}

	// 创建测试表 - 支持不同格式的 Vector
	_, err = db.Exec("CREATE TABLE vector_example(id INT, vec_f32 VECTOR(3, FLOAT32), vec_f64 VECTOR(4, FLOAT64))")
	if err != nil {
		fmt.Println("create table error:", err)
		return
	}
	fmt.Println("✓ Table created successfully")

	// ========================================
	// 1. 使用 Vector 结构体直接绑定插入
	// ========================================
	fmt.Println("\n=== Insert using Vector struct binding ===")

	// 插入 FLOAT32 向量
	vec1 := yasdb.Vector{
		Data:   []float32{1.0, 2.0, 3.0},
		Dim:    3,
		Format: yasdb.VectorFormatFloat32,
	}
	// 插入 FLOAT64 向量
	vec2 := yasdb.Vector{
		Data:   []float64{1.5, 2.5, 3.5, 4.5},
		Dim:    4,
		Format: yasdb.VectorFormatFloat64,
	}
	_, err = db.Exec("INSERT INTO vector_example VALUES(?, ?, ?)", 1, vec1, vec2)
	if err != nil {
		fmt.Println("insert error:", err)
		return
	}
	fmt.Printf("✓ Inserted Vector with float32: %v\n", vec1.Data)
	fmt.Printf("✓ Inserted Vector with float64: %v\n", vec2.Data)

	// ========================================
	// 2. 使用字符串形式插入（兼容旧方式）
	// ========================================
	fmt.Println("\n=== Insert using string format (legacy) ===")

	_, err = db.Exec("INSERT INTO vector_example VALUES(?, ?, ?)", 3, "[0.1, 0.2, 0.3]", "[0.1, 0.2, 0.3, 0.4]")
	if err != nil {
		fmt.Println("insert error:", err)
		return
	}
	fmt.Println("✓ Inserted using string format")

	// ========================================
	// 3. 查询数据
	// ========================================
	fmt.Println("\n=== Query Results ===")

	rows, err := db.Query("SELECT id, vec_f32, vec_f64 FROM vector_example ORDER BY id")
	if err != nil {
		fmt.Println("query error:", err)
		return
	}
	defer rows.Close()

	fmt.Println("ID | Vec_F32 (FLOAT32)        | Vec_F64 (FLOAT64)")
	fmt.Println("---+--------------------------+---------------------------")
	for rows.Next() {
		var id int
		var vecF32, vecF64 yasdb.Vector
		err := rows.Scan(&id, &vecF32, &vecF64)
		if err != nil {
			fmt.Println("scan error:", err)
			continue
		}
		fmt.Printf("%-3d| %-24v | %-25v\n", id, vecF32.Data, vecF64.Data)
	}

	// ========================================
	// 4. 使用 to_vector 函数查询
	// ========================================
	fmt.Println("\n=== Query with to_vector() function ===")

	rows, err = db.Query("SELECT id, to_vector(vec_f32) FROM vector_example ORDER BY id")
	if err != nil {
		fmt.Println("query error:", err)
		return
	}
	defer rows.Close()

	fmt.Println("ID | Vector (via to_vector)")
	fmt.Println("---+------------------------")
	for rows.Next() {
		var id int
		var vec yasdb.Vector
		err := rows.Scan(&id, &vec)
		if err != nil {
			fmt.Println("scan error:", err)
			continue
		}
		fmt.Printf("%-3d| %+v (dim=%d, format=%d)\n", id, vec.Data, vec.Dim, vec.Format)
	}

	// ========================================
	// 5. 使用 from_vector 转为字符串
	// ========================================
	fmt.Println("\n=== Query with from_vector() to string ===")

	rows, err = db.Query("SELECT id, from_vector(vec_f32) FROM vector_example ORDER BY id")
	if err != nil {
		fmt.Println("query error:", err)
		return
	}
	defer rows.Close()

	fmt.Println("ID | Vector (as string)")
	fmt.Println("---+--------------------")
	for rows.Next() {
		var id int
		var vecStr string
		err := rows.Scan(&id, &vecStr)
		if err != nil {
			fmt.Println("scan error:", err)
			continue
		}
		fmt.Printf("%-3d| %s\n", id, vecStr)
	}

	// ========================================
	// 6. 清理
	// ========================================
	_, err = db.Exec("DROP TABLE IF EXISTS vector_example")
	if err != nil {
		fmt.Println("cleanup error:", err)
	}
	fmt.Println("\n✓ Cleanup done")
}

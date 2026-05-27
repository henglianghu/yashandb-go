package main

import (
	"database/sql"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"

	_ "github.com/yashan-technologies/yashandb-go"

	example "github.com/yashan-technologies/yashandb-go/_examples"
)

var (
	dsn          string
	duration     time.Duration
	concurrency  int
	showMemStats bool
	tableName    = "stress_test_table"
)

type testResult struct {
	name      string
	success   int64
	fail      int64
	lastError string
	mu        sync.Mutex
}

func (r *testResult) addSuccess() {
	r.mu.Lock()
	r.success++
	r.mu.Unlock()
}

func (r *testResult) addFail(err error) {
	r.mu.Lock()
	r.fail++
	r.lastError = err.Error()
	r.mu.Unlock()
}

func main() {
	flag.DurationVar(&duration, "duration", 1*time.Hour, "test duration (e.g., 1h, 30m)")
	flag.IntVar(&concurrency, "concurrency", 1, "number of concurrent connections")
	flag.BoolVar(&showMemStats, "mem", true, "show memory stats periodically")
	flag.Parse()

	dsn = example.GetDsn()

	fmt.Println("===========================================")
	fmt.Println("YashanDB Stress Test")
	fmt.Println("===========================================")
	fmt.Printf("DSN:         %s\n", dsn)
	fmt.Printf("Duration:    %v\n", duration)
	fmt.Printf("Concurrency: %d\n", concurrency)
	fmt.Printf("Table:       %s\n", tableName)
	fmt.Println("===========================================")

	// 创建测试表
	if err := setupTable(); err != nil {
		fmt.Printf("Failed to setup table: %v\n", err)
		os.Exit(1)
	}

	// 启动测试
	startTime := time.Now()
	var wg sync.WaitGroup

	// 启动内存统计goroutine
	if showMemStats {
		go printMemStats()
	}

	// 启动多个并发测试
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go runStressTest(i, &wg)
	}

	// 等待结束
	wg.Wait()

	elapsed := time.Since(startTime)
	fmt.Println("\n===========================================")
	fmt.Println("Test Summary")
	fmt.Println("===========================================")
	fmt.Printf("Total time:  %v\n", elapsed)
	printMemStats()
	fmt.Println("===========================================")
}

// setupTable 创建测试表
func setupTable() error {
	db, err := sql.Open("yasdb", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// 删除已存在的表
	db.Exec("DROP TABLE IF EXISTS " + tableName)

	// 创建包含多种类型的表
	schema := fmt.Sprintf(`
		CREATE TABLE %s (
			id INT PRIMARY KEY,
			num_int INTEGER,
			num_bigint BIGINT,
			num_float FLOAT,
			num_double DOUBLE,
			num_decimal NUMBER,
			str_varchar VARCHAR(1000),
			str_char CHAR(100),
			dt_date DATE,
			dt_timestamp TIMESTAMP,
			dt_timestamp_tz TIMESTAMP WITH TIME ZONE,
			blob_data BLOB,
			clob_data CLOB,
			vec_vector VECTOR(3, FLOAT32)
		)
	`, tableName)

	_, err = db.Exec(schema)
	return err
}

// runStressTest 运行压力测试
func runStressTest(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	db, err := sql.Open("yasdb", dsn)
	if err != nil {
		fmt.Printf("[Worker %d] Failed to connect: %v\n", id, err)
		return
	}
	defer db.Close()

	// 设置连接池参数
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)

	endTime := time.Now().Add(duration)
	iteration := 0

	for time.Now().Before(endTime) {
		iteration++

		// 执行各类测试
		testInsert(db, id, iteration)
		testQuery(db, id, iteration)
		testQueryWithNull(db, id, iteration)
		testUpdate(db, id, iteration)
		testDeleteAndInsert(db, id, iteration)
		testTransaction(db, id, iteration)
		testLOB(db, id, iteration)
		testVector(db, id, iteration)
		testPrepareStatement(db, id, iteration)
		testBatch(db, id, iteration)

		// 随机休息一下，避免过度占用CPU
		if iteration%10 == 0 {
			time.Sleep(time.Millisecond * 10)
		}
	}

	fmt.Printf("[Worker %d] Finished after %d iterations\n", id, iteration)
}

// testInsert 测试插入
func testInsert(db *sql.DB, workerID, iter int) {
	// 生成随机数据
	id := workerID*1000000 + iter
	numInt := rand.Int31()
	numBigint := int64(rand.Int63())
	numFloat := rand.Float32()
	numDouble := rand.Float64()
	numDecimal := fmt.Sprintf("12345.%d", rand.Intn(10000))
	strVarchar := randomString(100)
	strChar := randomString(50)
	blobData := randomBytes(1000)
	clobData := randomString(500)
	vecData := fmt.Sprintf("[%f,%f,%f]", rand.Float32(), rand.Float32(), rand.Float32())

	query := fmt.Sprintf(`
		INSERT INTO %s (id, num_int, num_bigint, num_float, num_double, num_decimal,
			str_varchar, str_char, dt_date, dt_timestamp, dt_timestamp_tz,
			blob_data, clob_data, vec_vector)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, SYSDATE, SYSDATE, SYSDATE, ?, ?, TO_VECTOR(?))
	`, tableName)

	_, err := db.Exec(query, id, numInt, numBigint, numFloat, numDouble, numDecimal,
		strVarchar, strChar, blobData, clobData, vecData)

	if err != nil {
		fmt.Printf("[Worker %d] Insert failed: %v\n", workerID, err)
	}
}

// testQuery 测试查询
func testQuery(db *sql.DB, workerID, iter int) {
	id := workerID*1000000 + iter

	query := fmt.Sprintf("SELECT id, num_int, num_bigint, str_varchar, dt_date FROM %s WHERE id = ?", tableName)
	row := db.QueryRow(query, id)

	var rid int
	var numInt int32
	var numBigint int64
	var strVarchar string
	var dtDate interface{}

	err := row.Scan(&rid, &numInt, &numBigint, &strVarchar, &dtDate)
	if err != nil && err != sql.ErrNoRows {
		fmt.Printf("[Worker %d] Query failed: %v\n", workerID, err)
	}
}

// testQueryWithNull 测试含NULL的查询
func testQueryWithNull(db *sql.DB, workerID, iter int) {
	// 插入一条含NULL的数据
	id := workerID*1000000 + iter + 100000
	query := fmt.Sprintf("INSERT INTO %s (id, num_int, str_varchar) VALUES (?, NULL, NULL)", tableName)
	db.Exec(query, id)

	// 查询
	query = fmt.Sprintf("SELECT id, num_int, str_varchar FROM %s WHERE id = ?", tableName)
	row := db.QueryRow(query, id)

	var rid int
	var numInt sql.NullInt32
	var strVarchar sql.NullString

	err := row.Scan(&rid, &numInt, &strVarchar)
	if err != nil && err != sql.ErrNoRows {
		fmt.Printf("[Worker %d] QueryWithNull failed: %v\n", workerID, err)
	}
}

// testUpdate 测试更新
func testUpdate(db *sql.DB, workerID, iter int) {
	id := workerID*1000000 + iter
	newValue := randomString(50)

	query := fmt.Sprintf("UPDATE %s SET str_varchar = ? WHERE id = ?", tableName)
	result, err := db.Exec(query, newValue, id)
	if err != nil {
		fmt.Printf("[Worker %d] Update failed: %v\n", workerID, err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// 可能记录不存在，插入一条
		testInsert(db, workerID, iter)
	}
}

// testDeleteAndInsert 测试删除后插入
func testDeleteAndInsert(db *sql.DB, workerID, iter int) {
	id := workerID*1000000 + iter + 200000

	// 先删除
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)
	db.Exec(query, id)

	// 再插入
	query = fmt.Sprintf(`
		INSERT INTO %s (id, num_int, num_bigint, str_varchar)
		VALUES (?, ?, ?, ?)
	`, tableName)
	_, err := db.Exec(query, id, rand.Int31(), int64(rand.Int63()), randomString(50))
	if err != nil {
		fmt.Printf("[Worker %d] DeleteAndInsert failed: %v\n", workerID, err)
	}
}

// testTransaction 测试事务
func testTransaction(db *sql.DB, workerID, iter int) {
	id := workerID*1000000 + iter + 300000

	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	// 插入
	_, err = tx.Exec(fmt.Sprintf("INSERT INTO %s (id, num_int, str_varchar) VALUES (?, ?, ?)", tableName),
		id, rand.Int31(), randomString(50))
	if err != nil {
		return
	}

	// 查询
	row := tx.QueryRow(fmt.Sprintf("SELECT id FROM %s WHERE id = ?", tableName), id)
	var rid int
	row.Scan(&rid)

	// 提交
	tx.Commit()
}

// testLOB 测试LOB类型
func testLOB(db *sql.DB, workerID, iter int) {
	id := workerID*1000000 + iter + 400000

	blobData := randomBytes(5000)
	clobData := randomString(3000)

	query := fmt.Sprintf("INSERT INTO %s (id, blob_data, clob_data) VALUES (?, ?, ?)", tableName)
	_, err := db.Exec(query, id, blobData, clobData)
	if err != nil {
		fmt.Printf("[Worker %d] LOB insert failed: %v\n", workerID, err)
		return
	}

	// 查询LOB
	query = fmt.Sprintf("SELECT blob_data, clob_data FROM %s WHERE id = ?", tableName)
	row := db.QueryRow(query, id)

	var blob []byte
	var clob string
	err = row.Scan(&blob, &clob)
	if err != nil {
		fmt.Printf("[Worker %d] LOB query failed: %v\n", workerID, err)
	}
}

// testVector 测试Vector类型
func testVector(db *sql.DB, workerID, iter int) {
	id := workerID*1000000 + iter + 500000

	vecData := fmt.Sprintf("[%f,%f,%f]", rand.Float32(), rand.Float32(), rand.Float32())

	query := fmt.Sprintf("INSERT INTO %s (id, vec_vector) VALUES (?, TO_VECTOR(?))", tableName)
	_, err := db.Exec(query, id, vecData)
	if err != nil {
		fmt.Printf("[Worker %d] Vector insert failed: %v\n", workerID, err)
		return
	}

	// 查询Vector
	query = fmt.Sprintf("SELECT vec_vector FROM %s WHERE id = ?", tableName)
	row := db.QueryRow(query, id)

	var vec interface{}
	err = row.Scan(&vec)
	if err != nil {
		fmt.Printf("[Worker %d] Vector query failed: %v\n", workerID, err)
	}
}

// testPrepareStatement 测试预处理语句
func testPrepareStatement(db *sql.DB, workerID, iter int) {
	id := workerID*1000000 + iter + 600000

	stmt, err := db.Prepare(fmt.Sprintf("INSERT INTO %s (id, num_int, str_varchar) VALUES (?, ?, ?)", tableName))
	if err != nil {
		fmt.Printf("[Worker %d] Prepare failed: %v\n", workerID, err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(id, rand.Int31(), randomString(50))
	if err != nil {
		fmt.Printf("[Worker %d] Prepared statement exec failed: %v\n", workerID, err)
	}
}

// testBatch 测试批量操作
func testBatch(db *sql.DB, workerID, iter int) {
	id := workerID*1000000 + iter + 700000

	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (id, num_int) VALUES (?, ?)", tableName))
	if err != nil {
		return
	}
	defer stmt.Close()

	// 批量插入10条
	for i := 0; i < 10; i++ {
		stmt.Exec(id+i, rand.Int31())
	}

	tx.Commit()
}

// printMemStats 定期打印内存统计
func printMemStats() {
	var memStats runtime.MemStats
	prevAlloc := uint64(0)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		runtime.ReadMemStats(&memStats)

		allocMB := memStats.Alloc / 1024 / 1024
		totalAllocMB := memStats.TotalAlloc / 1024 / 1024
		sysMB := memStats.Sys / 1024 / 1024
		gcCount := memStats.NumGC

		// 计算增量
		delta := int64(allocMB) - int64(prevAlloc)
		deltaStr := ""
		if delta > 0 {
			deltaStr = fmt.Sprintf("(+%d MB)", delta)
		} else if delta < 0 {
			deltaStr = fmt.Sprintf("(%d MB)", delta)
		}

		fmt.Printf("[MemStats] Alloc: %d MB, TotalAlloc: %d MB, Sys: %d MB, GC: %d, Delta: %s\n",
			allocMB, totalAllocMB, sysMB, gcCount, deltaStr)

		prevAlloc = memStats.Alloc
	}
}

// randomString 生成随机字符串
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// randomBytes 生成随机字节数组
func randomBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(rand.Intn(256))
	}
	return b
}

package yasdb

import (
	"database/sql"
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	testDsn     = "sys/Cod-2022@127.0.0.1:1688"
	tablePrefix = "gosqltest_"
)

func init() {
	testing.Init()
	dsn := flag.String("dsn", "", "input you dsn(DataSourceName, format:username/password@host:port[?param1=value1&...&paramN=valueN]) to connect yashandb.")
	flag.Parse()
	if *dsn != "" {
		testDsn = *dsn
	}
	fmt.Println("test dsn:", testDsn)
}

type sqlTest struct {
	*sqlGenInfo
	*testing.T
	*sql.DB
}

type sqlGenInfo struct {
	tableName      string
	columnNameType [][2]string
	orderByColumn  string
	order          string
	execArgs       [][]interface{}
	queryArgs      []interface{}
	queryResult    [][]interface{}
	execResult     [][]interface{}
	query          string
	DBVersion      string
}

func newSqlTest(t *testing.T) *sqlTest {
	db, err := sql.Open("yasdb", testDsn)
	if err != nil {
		t.Fatalf("%s%v", NormalConnErr, err)
	}
	return &sqlTest{T: t, DB: db}
}

func newSqlAutoCommitTest(t *testing.T) *sqlTest {
	db, err := sql.Open("yasdb", fmt.Sprintf("%s?%s", testDsn, "autocommit=true"))
	if err != nil {
		t.Fatalf("%s%v", NormalConnErr, err)
	}
	return &sqlTest{T: t, DB: db}
}

func (st *sqlTest) getDBVersion() {
	r, err := st.DB.Query("select version_number from v$version")
	if err != nil {
		st.T.Fatalf("get db version failed %s", err)
	}
	if r.Next() {
		version := ""
		if err := r.Scan(&version); err != nil {
			st.T.Fatalf("scan db version failed %s", err)
		}
		st.DBVersion = version
	}
}

func (st *sqlTest) IsYasdbMode() bool {
	r, err := st.DB.Query("select VALUE from v$parameter where name = 'EMPTY_STRING_AS_NULL'")
	if err != nil {
		return true
	}
	yasMode := true
	if r.Next() {
		_ = r.Scan(&yasMode) // EMPTY_STRING_AS_NULL = FALSE  is mysql mode
	}
	return yasMode
}

func (st *sqlTest) isUdtXmltype(columnType string) bool {
	if st.DBVersion == "" {
		st.getDBVersion()
	}
	if strings.ToLower(columnType) != "xmltype" {
		return false
	}
	c, err := CompareVersion(st.DBVersion, "23.4.2.100")
	if err != nil {
		return false
	}
	return c >= 0
}

func (st *sqlTest) isToTimestampTzSupport() bool {
	if st.DBVersion == "" {
		st.getDBVersion()
	}
	c, err := CompareVersion(st.DBVersion, "23.4.2.100")
	if err != nil {
		return false
	}
	return c >= 0
}

func (st *sqlTest) isBfileSupport() bool {
	if st.DBVersion == "" {
		st.getDBVersion()
	}
	c, err := CompareVersion(st.DBVersion, "23.4.2.0")
	if err != nil {
		return false
	}
	return c >= 0
}

func (st *sqlTest) dropTable() {
	dropTableSql := fmt.Sprintf("drop table if exists %s", st.tableName)
	st.mustExec(dropTableSql)
}

func (st *sqlTest) createTable() {
	columnDefineStr := ""
	flag := false
	for _, column := range st.columnNameType {
		if flag {
			columnDefineStr += fmt.Sprintf(", %s %s", column[0], column[1])
		} else {
			columnDefineStr += fmt.Sprintf("%s %s", column[0], column[1])
		}
		flag = true
	}
	createTableSql := fmt.Sprintf(`create table %s( %s )`, st.tableName, columnDefineStr)
	st.mustExec(createTableSql)
}

func (st *sqlTest) genTableTest() {
	st.dropTable()
	st.createTable()
}

func (st *sqlTest) mustExec(query string, args ...interface{}) sql.Result {
	res, err := st.Exec(query, args...)
	if err != nil {
		st.Fatalf("Error running %s [%v]: %v\n", query, args, err)
	}
	return res
}

func (st *sqlTest) mustQuery(query string, args ...interface{}) *sql.Rows {
	res, err := st.Query(query, args...)
	if err != nil {
		st.Fatalf("Error running %s [%v]: %v\n", query, args, err)
	}
	return res
}

func (st *sqlTest) runInsertTest() {
	columnName := ""
	bindValue := ""
	flag := false
	for _, column := range st.columnNameType {

		v := "?"
		if st.isUdtXmltype(column[1]) {
			v = "xmltype(?)"
		}

		if flag {
			columnName += ", " + column[0]
			bindValue += fmt.Sprintf(", %s", v)

		} else {
			columnName += column[0]
			bindValue += v
		}
		flag = true
	}
	st.query = fmt.Sprintf("insert into %s (%s) values(%s)", st.tableName, columnName, bindValue)
	st.runExecTest()
}

func (st *sqlTest) runExecTest() {
	for i := range st.execArgs {
		st.mustExec(st.query, st.execArgs[i]...)
	}
}

func (st *sqlTest) runSelectTest() {
	columnName := ""
	flag := false
	for _, column := range st.columnNameType {
		colName := column[0]
		if st.isUdtXmltype(column[1]) {
			colName = fmt.Sprintf("%s.%s.getClobVal()", st.tableName, column[0])
		}
		if flag {
			columnName += ", " + colName
		} else {
			columnName += colName
		}
		flag = true
	}
	st.query = fmt.Sprintf("select %s from %s %s", columnName, st.tableName, st.genOrderBy())
	st.runQueryTest()
}

func (st *sqlTest) runQueryTest() {
	actualResult, err := st.getRowsTest()
	if err != nil {
		st.Errorf("get rows error: %v - query: %s", err, st.query)
		return
	}
	if err = st.resultComparison(actualResult, st.queryResult); err != nil {
		st.Errorf("result comparison error: %v - query: %s", err, st.query)
	}
}

func (st *sqlTest) genOrderBy() string {
	orderBy := "order by "
	if st.orderByColumn == "" {
		orderBy += "1"
	} else {
		orderBy += fmt.Sprintf("%s %s", st.orderByColumn, st.order)
	}
	return orderBy
}

func (st *sqlTest) getRowsTest() ([][]interface{}, error) {
	rows := st.mustQuery(st.query, st.queryArgs...)
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		rows.Close()
		return nil, fmt.Errorf("columns error: %v", err)
	}

	result := make([][]interface{}, 0)
	columnCount := len(columns)
	rowResult := make([]interface{}, columnCount)
	for rows.Next() {
		rowInterface := make([]interface{}, columnCount)
		for i := 0; i < columnCount; i++ {
			rowResult[i] = &rowInterface[i]
		}
		err = rows.Scan(rowResult...)
		if err != nil {
			return nil, fmt.Errorf("scan error: %v", err)
		}
		result = append(result, rowInterface)
	}

	return result, nil
}

func (st *sqlTest) getRowsColumnTypes() ([]*sql.ColumnType, error) {
	rows := st.mustQuery(st.query, st.queryArgs...)
	defer rows.Close()

	return rows.ColumnTypes()
}

func (st *sqlTest) resultComparison(actualResults [][]interface{}, expectedResults [][]interface{}) error {
	if actualResults == nil && expectedResults != nil {
		return fmt.Errorf("result is nil")
	}
	if len(actualResults) != len(expectedResults) {
		return fmt.Errorf("result rows len %v not equal to expected results len %v",
			len(actualResults), len(expectedResults))
	}

	testTypeTime := reflect.TypeOf(time.Time{})
	for i := 0; i < len(actualResults); i++ {
		for j := 0; j < len(actualResults[i]); j++ {
			actualResult := actualResults[i][j]
			expectedResult := expectedResults[i][j]
			bad := false
			type1 := reflect.TypeOf(actualResult)
			type2 := reflect.TypeOf(expectedResult)
			switch {
			case type1 == nil || type2 == nil:
				if type1 != type2 {
					bad = true
				}
			case type1 == testTypeTime || type2 == testTypeTime:
				if type1 != type2 {
					bad = true
					break
				}
				time1 := actualResult.(time.Time)
				time2 := expectedResult.(time.Time)
				if !time1.Equal(time2) {
					bad = true
				}
			case type1.Kind() == reflect.Slice || type2.Kind() == reflect.Slice:
				if !reflect.DeepEqual(actualResult, expectedResult) {
					bad = true
				}
			default:
				if actualResult != expectedResult {
					bad = true
				}
			}
			if bad {
				return fmt.Errorf("result - row %v, %v - received: %T, %v - expected: %T, %v",
					i, j, actualResult, actualResult, expectedResult, expectedResult)
			}

		}
	}

	return nil
}

func runSqlTest(t *testing.T, fn func(st *sqlTest)) {
	db, err := sql.Open("yasdb", testDsn)
	if err != nil {
		t.Fatalf("%s%v", NormalConnErr, err)
	}
	defer db.Close()
	fn(&sqlTest{T: t, DB: db})
}

func runsqlTestACTrue(t *testing.T, fn func(st *sqlTest)) {
	dsn := fmt.Sprintf("%s?%s", testDsn, "autocommit=true")
	db, err := sql.Open("yasdb", dsn)
	if err != nil {
		t.Fatalf("%s%v", NormalConnErr, err)
	}
	defer db.Close()
	fn(&sqlTest{T: t, DB: db})
}

func runsqlTestACFalse(t *testing.T, fn func(st *sqlTest)) {
	dsn := fmt.Sprintf("%s?%s", testDsn, "autocommit=false")
	db, err := sql.Open("yasdb", dsn)
	if err != nil {
		t.Fatalf("%s%v", NormalConnErr, err)
	}
	defer db.Close()
	fn(&sqlTest{T: t, DB: db})
}

func affectedResultComparison(t *testing.T, actualResult int64, expectedResult int64) {
	if actualResult != expectedResult {
		t.Fatalf("affected result - received: %d - expected: %d", actualResult, expectedResult)
	}
}

var (
	VERSION_LEN = 4
)

func CompareVersionWithoutSuf(v1, v2 string) (int, error) {
	shortV1 := strings.Split(v1, "-")[0]
	shortV2 := strings.Split(v2, "-")[0]
	return CompareVersion(shortV1, shortV2)
}

// CompareVersion 比较版本
// if v1 > v2, return 1
// if v1 = v2, return 0
// if v1 < v2, return -1
func CompareVersion(v1, v2 string) (int, error) {
	v1List := strings.Split(v1, ".")
	v2List := strings.Split(v2, ".")

	// 版本是xx.xx.xx.xx
	if len(v1List) != VERSION_LEN || len(v2List) != VERSION_LEN {
		return 0, fmt.Errorf("version format error")
	}

	// 逐位比较
	for i := 0; i < VERSION_LEN; i++ {
		first, err := strconv.Atoi(v1List[i])
		if err != nil {
			return 0, err
		}
		second, err := strconv.Atoi(v2List[i])
		if err != nil {
			return 0, err
		}
		if first > second {
			return 1, nil
		} else if first < second {
			return -1, nil
		}
	}
	// 四位均相等
	return 0, nil
}

package yasdb

import (
	"testing"
	"time"
)

var (
	_TimeFormat      = "15:04:05.000000"
	_TimestampForamt = "2006-01-02 15:04:05.000000"
	_DateFormat      = "2006-01-02 15:04:05"
)

func TestDate(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testDate)
}

func testDate(t *sqlTest) {
	// DATE；取值范围：0001-01-01 00:00:00 ~ 9999-12-31 23:59:59；精度:秒
	si := sqlGenInfo{}
	t.sqlGenInfo = &si
	t1 := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	t3 := time.Unix(time.Now().Unix(), 0)
	t3utc := time.Date(t3.Year(), t3.Month(), t3.Day(), t3.Hour(), t3.Minute(), t3.Second(), 0, time.UTC)
	t4, _ := time.Parse(_DateFormat, "0001-01-01 00:00:00")
	si = sqlGenInfo{
		tableName: "test_date",
		columnNameType: [][2]string{
			{"c1", "INT"},
			{"c2", "DATE"},
		},
		execArgs: [][]interface{}{
			{1, t1},
			{2, t2},
			{3, t3},
			{4, t4},
		},
		queryResult: [][]interface{}{
			{int32(1), t1},
			{int32(2), t2},
			{int32(3), t3utc},
			{int32(4), t4},
		},
	}
	t.genTableTest()
	t.runInsertTest()
	t.runSelectTest()
}

func TestTime(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testTime)
}

func testTime(t *sqlTest) {
	// TIME；取值范围：00:00:00.000000 ~ 23:59:59.999999；精度: 微秒
	si := sqlGenInfo{}
	t.sqlGenInfo = &si
	t1, _ := time.Parse(_TimeFormat, "00:00:00.000000")
	t2, _ := time.Parse(_TimeFormat, "23:59:59.999999")
	t3, _ := time.Parse(_TimeFormat, "12:01:10.121212")
	si = sqlGenInfo{
		tableName: "test_time",
		columnNameType: [][2]string{
			{"c1", "INT"},
			{"c2", "TIME"},
		},
		execArgs: [][]interface{}{
			{1, t1},
			{2, t2},
			{3, t3},
		},
		queryResult: [][]interface{}{
			{int32(1), t1},
			{int32(2), t2},
			{int32(3), t3},
		},
	}
	t.genTableTest()
	t.runInsertTest()
	t.runSelectTest()
}

func TestTimestamp(t *testing.T) {
	t.Parallel()
	runSqlTest(t, testTimestamp)
}

func testTimestamp(t *sqlTest) {
	// TIMESTAMP；取值范围：1-1-1 00:00:00.000000 ~ 9999-12-31 23:59:59.999999；精度: 微秒
	si := sqlGenInfo{}
	t.sqlGenInfo = &si
	t1, _ := time.Parse(_TimestampForamt, "0001-01-01 00:00:00.000000")
	t2, _ := time.Parse(_TimestampForamt, "9999-12-31 23:59:59.999999")
	t3 := time.Unix(0, time.Now().UnixMilli()*1000)
	t3utc := time.Date(t3.Year(), t3.Month(), t3.Day(), t3.Hour(), t3.Minute(), t3.Second(), t3.Nanosecond(), time.UTC)
	si = sqlGenInfo{
		tableName: "test_timestamp",
		columnNameType: [][2]string{
			{"c1", "INT"},
			{"c2", "TIMESTAMP"},
		},
		execArgs: [][]interface{}{
			{1, t1},
			{2, t2},
			{3, t3},
		},
		queryResult: [][]interface{}{
			{int32(1), t1},
			{int32(2), t2},
			{int32(3), t3utc},
		},
	}
	t.genTableTest()
	t.runInsertTest()
	t.runSelectTest()
}

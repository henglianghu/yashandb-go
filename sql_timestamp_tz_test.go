package yasdb

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestTimestampTimezoneTypes(t *testing.T) {
	runSqlTest(t, testTimestampTimezoneTypes)
}

func testTimestampTimezoneTypes(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	if !t.isToTimestampTzSupport() {
		t.Skip("database does not support TIMESTAMP WITH TIME ZONE")
	}

	tableName := "test_timestamp_tz_types"
	t.mustExec("drop table if exists " + tableName)
	t.mustExec(fmt.Sprintf(`
		CREATE TABLE %s (
			id     INT PRIMARY KEY,
			ts     TIMESTAMP,
			ts_tz  TIMESTAMP WITH TIME ZONE,
			ts_ltz TIMESTAMP WITH LOCAL TIME ZONE
		)`, tableName))
	defer t.mustExec("drop table " + tableName)

	// 驱动通过 UnixMicro() 传入绝对时间（UTC 微秒值）
	// 数据库将其解释为 UTC 时间
	cases := []struct {
		id int
		ts time.Time // 用于比较 Unix 时间戳
	}{
		{
			id: 1,
			ts: time.Date(2024, 6, 15, 10, 30, 45, 123456000, time.UTC),
		},
		{
			id: 2,
			ts: time.Date(2024, 12, 25, 8, 0, 0, 0, time.UTC),
		},
		{
			id: 3,
			ts: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		},
	}

	for _, c := range cases {
		t.mustExec(
			fmt.Sprintf("insert into %s(id, ts, ts_tz, ts_ltz) values(?,?,?,?)", tableName),
			c.id, c.ts, c.ts, c.ts,
		)
	}

	rows, err := t.DB.Query(fmt.Sprintf("select id, ts, ts_tz, ts_ltz from %s order by id", tableName))
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	rowCount := 0
	for rows.Next() {
		var (
			id    int32
			ts    time.Time
			tsTz  time.Time
			tsLtz time.Time
		)

		if err := rows.Scan(&id, &ts, &tsTz, &tsLtz); err != nil {
			t.Fatalf("row %d scan failed: %v", rowCount+1, err)
		}
		expected := cases[rowCount]
		if id != int32(expected.id) {
			t.Fatalf("row %d: id got %d, want %d", rowCount+1, id, expected.id)
		}
		_TimeZoneLayout := DateTimeFormats[DateTimeMicroZone]
		fmt.Println(ts.Format(_TimeZoneLayout), tsTz.Format(_TimeZoneLayout), tsLtz.Format(_TimeZoneLayout))
		fmt.Println(expected.ts.Format(_TimeZoneLayout))

		// TIMESTAMP / TIMESTAMP WITH TIME ZONE / TIMESTAMP WITH LOCAL TIME ZONE
		// 驱动以 UnixMicro() 传入 UTC 微秒值
		// TIMESTAMP 和 TIMESTAMP_TZ：返回 UTC wall clock，比较 Unix 时间戳
		// TIMESTAMP_LTZ：返回 DB TZ 转换后的 wall clock，比较 Unix 时间戳
		// 所有类型在绝对时间上应保持一致
		assertTimestampEqual(t, ts, expected.ts, "ts")
		assertTimestampEqual(t, tsTz, expected.ts, "ts_tz")
		assertTimestampEqual(t, tsLtz, expected.ts, "ts_ltz")

		rowCount++
	}

	if rowCount != len(cases) {
		t.Fatalf("expected %d rows, got %d", len(cases), rowCount)
	}
}

func TestTimestampTimezoneInsertAndQuery(t *testing.T) {
	runSqlTest(t, testTimestampTimezoneInsertAndQuery)
}

func testTimestampTimezoneInsertAndQuery(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	if !t.isToTimestampTzSupport() {
		t.Skip("database does not support TIMESTAMP WITH TIME ZONE")
	}

	tableName := "test_timestamp_tz_iq"
	t.mustExec("drop table if exists " + tableName)
	t.mustExec(fmt.Sprintf(`
		CREATE TABLE %s (
			id    INT PRIMARY KEY,
			ts_tz  TIMESTAMP WITH TIME ZONE,
			ts_ltz TIMESTAMP WITH LOCAL TIME ZONE
		)`, tableName))
	defer t.mustExec("drop table " + tableName)

	insertCases := []struct {
		id    int
		tsTz  time.Time
		tsLtz time.Time
	}{
		{id: 1, tsTz: time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC), tsLtz: time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)},
		{id: 2, tsTz: time.Date(2024, 7, 20, 23, 59, 59, 999999000, time.UTC), tsLtz: time.Date(2024, 7, 20, 23, 59, 59, 999999000, time.UTC)},
		{id: 3, tsTz: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), tsLtz: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	for _, c := range insertCases {
		t.mustExec(
			fmt.Sprintf("insert into %s(id, ts_tz, ts_ltz) values(?,?,?)", tableName),
			c.id, c.tsTz, c.tsLtz,
		)
	}

	for _, c := range insertCases {
		var tsTz, tsLtz time.Time
		err := t.DB.QueryRow(fmt.Sprintf("select ts_tz, ts_ltz from %s where id = ?", tableName), c.id).Scan(&tsTz, &tsLtz)
		if err != nil {
			t.Fatal(err)
		}

		assertTimestampEqual(t, tsTz, c.tsTz, fmt.Sprintf("ts_tz id=%d", c.id))
		assertTimestampEqual(t, tsLtz, c.tsLtz, fmt.Sprintf("ts_ltz id=%d", c.id))
	}
}

func TestTimestampTimezoneReturningInto(t *testing.T) {
	runSqlTest(t, testTimestampTimezoneReturningInto)
}

func testTimestampTimezoneReturningInto(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	if !t.isToTimestampTzSupport() {
		t.Skip("database does not support TIMESTAMP WITH TIME ZONE")
	}

	tableName := "test_timestamp_tz_returning"
	t.mustExec("drop table if exists " + tableName)
	t.mustExec(fmt.Sprintf(`
		CREATE TABLE %s (
			id     INT,
			ts     TIMESTAMP,
			ts_tz  TIMESTAMP WITH TIME ZONE,
			ts_ltz TIMESTAMP WITH LOCAL TIME ZONE
		)`, tableName))
	defer t.mustExec("drop table " + tableName)

	now := time.Date(2024, 7, 1, 12, 0, 0, 0, time.UTC)

	var outTs, outTsTz, outTsLtz time.Time
	t.mustExec(
		fmt.Sprintf(`insert into %s(id, ts, ts_tz, ts_ltz) values(?,?,?,?)
		returning ts, ts_tz, ts_ltz into ?,?,?`, tableName),
		1, now, now, now,
		sql.Out{Dest: &outTs},
		sql.Out{Dest: &outTsTz},
		sql.Out{Dest: &outTsLtz},
	)

	assertTimestampEqual(t, outTs, now, "returning ts")
	assertTimestampEqual(t, outTsTz, now, "returning ts_tz")
	assertTimestampEqual(t, outTsLtz, now, "returning ts_ltz")
}

func TestTimestampTimezoneEdgeValues(t *testing.T) {
	runSqlTest(t, testTimestampTimezoneEdgeValues)
}

func testTimestampTimezoneEdgeValues(t *sqlTest) {
	t.sqlGenInfo = &sqlGenInfo{}
	if !t.isToTimestampTzSupport() {
		t.Skip("database does not support TIMESTAMP WITH TIME ZONE")
	}

	tableName := "test_timestamp_tz_edge"
	t.mustExec("drop table if exists " + tableName)
	t.mustExec(fmt.Sprintf(`
		CREATE TABLE %s (
			id     INT PRIMARY KEY,
			ts     TIMESTAMP,
			ts_tz  TIMESTAMP WITH TIME ZONE,
			ts_ltz TIMESTAMP WITH LOCAL TIME ZONE
		)`, tableName))
	defer t.mustExec("drop table " + tableName)

	cases := []struct {
		name string
		ts   time.Time
	}{
		{"normal", time.Date(2024, 6, 15, 10, 30, 45, 0, time.UTC)},
		{"epoch", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"y2k", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	for _, c := range cases {
		t.mustExec(
			fmt.Sprintf("insert into %s(id, ts, ts_tz, ts_ltz) values(?,?,?,?)", tableName),
			1, c.ts, c.ts, c.ts,
		)

		var ts, tsTz, tsLtz time.Time
		err := t.DB.QueryRow(fmt.Sprintf("select ts, ts_tz, ts_ltz from %s where id = 1", tableName)).Scan(&ts, &tsTz, &tsLtz)
		if err != nil {
			t.Fatalf("case %s: scan failed: %v", c.name, err)
		}

		assertTimestampEqual(t, ts, c.ts, fmt.Sprintf("%s ts", c.name))
		assertTimestampEqual(t, tsTz, c.ts, fmt.Sprintf("%s ts_tz", c.name))
		assertTimestampEqual(t, tsLtz, c.ts, fmt.Sprintf("%s ts_ltz", c.name))

		t.mustExec(fmt.Sprintf("delete from %s where id = 1", tableName))
	}
}

func assertTimestampEqual(t testing.TB, got, want time.Time, label string) {
	// 比较 wall clock（年月日时分秒），而非 Unix 时间戳
	// 因为数据库可能将值转换为不同的时区
	if got.Year() != want.Year() ||
		got.Month() != want.Month() ||
		got.Day() != want.Day() ||
		got.Hour() != want.Hour() ||
		got.Minute() != want.Minute() ||
		got.Second() != want.Second() ||
		got.Nanosecond() != want.Nanosecond() {
		t.Fatalf("%s: wall clock mismatch: got %v, want %v", label, got.Format("2006-01-02 15:04:05.000000"), want.Format("2006-01-02 15:04:05.000000"))
	}
}

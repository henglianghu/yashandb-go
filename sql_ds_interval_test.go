package yasdb

import (
	"testing"
	"time"
)

func TestDSInterval(t *testing.T) {
	runSqlTest(t, testDsInterval)
}

func testDsInterval(t *sqlTest) {
	si := sqlGenInfo{}
	t.sqlGenInfo = &si

	si = sqlGenInfo{
		tableName: "test_ds_interval",
		columnNameType: [][2]string{
			{"id", "int"},
			{"c1", "INTERVAL DAY(9) TO SECOND(9)"},
		},
		execArgs: [][]interface{}{
			{1, nil},
			{2, DSInterval{Day: 100000000, DayTime: time.Date(0, 0, 0, 23, 23, 23, 999999*1e3, time.UTC)}},
			{3, DSInterval{Day: -100000000, DayTime: time.Date(0, 0, 0, 23, 23, 23, 999999*1e3, time.UTC)}},
		},
		queryResult: [][]interface{}{
			{int32(1), nil},
			{int32(2), DSInterval{Day: 100000000, DayTime: time.Date(0, 0, 0, 23, 23, 23, 999999*1e3, time.UTC)}},
			{int32(3), DSInterval{Day: -100000000, DayTime: time.Date(0, 0, 0, 23, 23, 23, 999999*1e3, time.UTC)}},
		},
	}
	t.genTableTest()
	t.runInsertTest()
	t.runSelectTest()
}

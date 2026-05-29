package yasdb

import "testing"

func TestYmInterval(t *testing.T) {
	runSqlTest(t, testYmInterval)
}

func testYmInterval(t *sqlTest) {
	si := sqlGenInfo{}
	t.sqlGenInfo = &si

	si = sqlGenInfo{
		tableName: "test_ym_interval",
		columnNameType: [][2]string{
			{"id", "int"},
			{"c1", "interval year(9) to month"},
		},
		execArgs: [][]interface{}{
			{1, nil},
			{2, YMInterval{Year: 178000000, Month: 10}},
			{3, YMInterval{Year: -178000000, Month: 10}},
		},
		queryResult: [][]interface{}{
			{int32(1), nil},
			{int32(2), YMInterval{Year: 178000000, Month: 10}},
			{int32(3), YMInterval{Year: -178000000, Month: 10}},
		},
	}
	t.genTableTest()
	t.runInsertTest()
	t.runSelectTest()
}

package yasdb

import "testing"

func TestXml(t *testing.T) {
	runSqlTest(t, testXml)
}

func testXml(t *sqlTest) {
	si := sqlGenInfo{}
	t.sqlGenInfo = &si

	si = sqlGenInfo{
		tableName: "test_xml",
		columnNameType: [][2]string{
			{"id", "int"},
			{"c1", "xmltype"}, // xmltype改成了UDT，不好搞
		},
		execArgs: [][]interface{}{
			{1, "<test></test>"},
			{2, "<data>sics</data>"},
		},
		queryResult: [][]interface{}{
			{int32(1), "<test></test>"},
			{int32(2), "<data>sics</data>"},
		},
	}
	t.genTableTest()
	t.runInsertTest()
	t.runSelectTest()
}

package yasdb

import (
	"testing"
)

func TestDatabaseTypeName(t *testing.T) {
	runSqlTest(t, testNumberFloatDatabaseTypeName)
}

func testNumberFloatDatabaseTypeName(t *sqlTest) {
	si := sqlGenInfo{}
	t.sqlGenInfo = &si

	columnNameType := [][2]string{
		{"c1", "TINYINT"},
		{"c2", "SMALLINT"},
		{"c3", "INT"},
		{"c4", "BIGINT"},
		{"c5", "FLOAT"},
		{"c6", "DOUBLE"},
		{"c7", "NUMBER"},
		{"c8", "BIT(64)"},
		{"c9", "CHAR(126)"},
		{"c10", "VARCHAR(126)"},
		{"c11", "NCHAR(126)"},
		{"c12", "NVARCHAR(126)"},
		{"c13", "BOOLEAN"},
		{"c14", "TIME"},
		{"c15", "TIMESTAMP"},
		{"c18", "INTERVAL YEAR TO MONTH"},
		{"c19", "INTERVAL DAY TO SECOND"},
		{"c20", "BLOB"},
		{"c21", "CLOB"},
		{"c22", "NCLOB"},
		{"c23", "RAW(10)"},
		{"c24", "JSON"},
		{"c26", "ROWID"},
		{"c27", "UROWID(20)"},
	}

	columnTypes := []string{
		"TINYINT",
		"SMALLINT",
		"INTEGER",
		"BIGINT",
		"FLOAT",
		"DOUBLE",
		"NUMBER",
		"BIT",
		"CHAR",
		"VARCHAR",
		"NCHAR",
		"NVARCHAR",
		"BOOLEAN",
		"TIME",
		"TIMESTAMP",
		"INTERVAL YEAR TO MONTH",
		"INTERVAL DAY TO SECOND",
		"BLOB",
		"CLOB",
		"NCLOB",
		"RAW",
		"JSON",
		"ROWID",
		"RAW",
	}

	if t.isBfileSupport() {
		columnNameType = append(columnNameType, [2]string{"c25", "BFILE"})
		columnTypes = append(columnTypes, "BFILE")
	}
	if t.isToTimestampTzSupport() {
		columnNameType = append(columnNameType, [][2]string{
			{"c16", "TIMESTAMP WITH LOCAL TIME ZONE"},
			{"c17", "TIMESTAMP WITH TIME ZONE"}}...)
		columnTypes = append(columnTypes,
			"TIMESTAMP WITH LOCAL TIME ZONE",
			"TIMESTAMP WITH TIME ZONE")
	}

	si = sqlGenInfo{
		tableName:      "database_type_name",
		columnNameType: columnNameType,
	}
	t.genTableTest()
	t.runSelectTest()

	rows, err := t.getRowsColumnTypes()
	if err != nil {
		t.Fatalf(err.Error())
	}
	for i, row := range rows {
		if row.DatabaseTypeName() != columnTypes[i] {
			t.Fatalf("column %d database type expected: %s actual: %s", i, columnTypes[i], row.DatabaseTypeName())
		}
	}
}

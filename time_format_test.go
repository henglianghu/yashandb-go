package yasdb

import (
	"testing"
	"time"

	"github.com/yashan-technologies/yashandb-go/assert"
)

func TestFormatTime(t *testing.T) {

	testDate := time.Date(2001, 5, 5, 14, 23, 23, 0, time.Local)

	cases := []struct {
		format   string
		expected string
	}{
		{
			format:   "YYYY-MM-DD",
			expected: "2001-05-05",
		},
		{
			format:   "YYYY/MM/DD",
			expected: "2001/05/05",
		},
		{
			format:   "YY/MM/DD",
			expected: "01/05/05",
		},
		{
			format:   "YYY-MM-DD",
			expected: "001-05-05",
		},
		// fixme: 不一致
		// {
		// 	format:   "Y-MM-DD",
		// 	expected: "1-05-05",
		// },
		{
			format:   "MON,DD,YYYY",
			expected: "May,05,2001",
		},
		{
			format:   "MONTH,DD,YYYY",
			expected: "May,05,2001",
		},
		{
			format:   "HH:MI:SS",
			expected: "14:23:23",
		},
		{
			format:   "HH24:MI:SS",
			expected: "14:23:23",
		},
		{
			format:   "HH12:MI:SS AM",
			expected: "02:23:23 PM",
		},
		{
			format:   "HH12:MI:SS P.M.",
			expected: "02:23:23 P.M.",
		},
		{
			format:   "HH12:MI:SS P.M.",
			expected: "02:23:23 P.M.",
		},
		{
			format:   "HH12:MI:SS A.M.",
			expected: "02:23:23 P.M.",
		},
		{
			format:   "YYYY/MM/DD HH24:MI:SS.FF",
			expected: "2001/05/05 14:23:23.000000",
		},
		{
			format:   "YYYY/MM/DD HH24:MI:SS.FF3",
			expected: "2001/05/05 14:23:23.000",
		},
		{
			format:   "YYYY-MM-DD HH24:MI:SS.FF TZH:TZM",
			expected: "2001-05-05 14:23:23.000000 +08:00",
		},
		{
			format:   "YYYY-MM-D HH24:MI:SS.FF TZH:TZM",
			expected: "2001-05-5 14:23:23.000000 +08:00",
		},
		{
			format:   "YYYY-MM-DD HH24:MI:SS.FF TZH:TZM",
			expected: "2001-05-05 14:23:23.000000 +08:00",
		},
	}

	assert := assert.NewAssert(t)
	for _, c := range cases {
		actual := FormatTime(c.format, testDate)
		assert.Equal(actual, c.expected)
	}

}

func TestFormatYMInterval(t *testing.T) {

	cases := []struct {
		year     int32
		month    int32
		format   string
		expected string
	}{
		{
			year:     1,
			month:    0,
			format:   "YY-MM",
			expected: "01-00",
		},
		{
			year:     -1,
			month:    11,
			format:   "YY-MM",
			expected: "-01-11",
		},
		{
			year:     -1,
			month:    11,
			format:   "YYYY-MM",
			expected: "-0001-11",
		},
		{
			year:     -1,
			month:    11,
			format:   "YYYY/MM",
			expected: "-0001/11",
		},
		{
			year:     2,
			month:    11,
			format:   "MM/YYYY",
			expected: "11/0002",
		},
		{
			year:     999999,
			month:    11,
			format:   "MM/YYYY",
			expected: "11/999999",
		},
	}

	assert := assert.NewAssert(t)
	for _, c := range cases {
		actual := FormatYMInterval(c.format, c.year, c.month)
		assert.Equal(actual, c.expected)
	}

}

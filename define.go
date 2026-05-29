/*
Copyright  2022, YashanDB and/or its affiliates. All rights reserved.
YashanDB Driver for golang is licensed under the terms of the mulan PSL v2.0

License: 	http://license.coscl.org.cn/MulanPSL2
Home page: 	https://www.yashandb.com/
*/
package yasdb

import "time"

// https://cod-doc.yasdb.com/yashandb/23.4alpha/zh/All-Manuals/Development-Guide/SQL-Reference-Manual/Data-Types/Date-Time-Types.html#%E5%B9%B4%E5%88%B0%E6%9C%88%E9%97%B4%E9%9A%94%E7%B1%BB%E5%9E%8B-interval-year-to-month-data-type
type YMInterval struct {
	Year  int32
	Month int32 // 有效值 为 [0,11]
}

// https://cod-doc.yasdb.com/yashandb/23.4alpha/zh/All-Manuals/Development-Guide/SQL-Reference-Manual/Data-Types/Date-Time-Types.html#%E5%A4%A9%E5%88%B0%E7%A7%92%E9%97%B4%E9%9A%94%E7%B1%BB%E5%9E%8B-interval-day-to-second-data-type
type DSInterval struct {
	Day     int32
	DayTime time.Time // 只取time中的hour,minute,second,Nanosecond(实际上只到微妙)
}

func NewYMInterval(year int32, month int32) YMInterval {
	return YMInterval{
		Year:  year,
		Month: month,
	}
}

func NewDSInterval(day int32, dayTime time.Time) DSInterval {
	return DSInterval{
		Day:     day,
		DayTime: dayTime,
	}
}

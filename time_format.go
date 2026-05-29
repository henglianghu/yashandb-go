package yasdb

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type DTF int8

const (
	DateTimeMicroZone DTF = iota
	DateTimeZone
	DateTimeMicro
	DateTime
	Date
	TimeMicro
	Time
)

// DateTimeFormats holds all datetime format strings
var DateTimeFormats = map[DTF]string{
	DateTimeMicroZone: "2006-01-02 15:04:05.999999 -07:00",
	DateTimeZone:      "2006-01-02 15:04:05 -07:00",
	DateTimeMicro:     "2006-01-02 15:04:05.999999",
	DateTime:          "2006-01-02 15:04:05",
	Date:              "2006-01-02",
	TimeMicro:         "15:04:05.999999",
	Time:              "15:04:05",
}

/*
取值范围/格式：[0001-01-01 00:00:00.000000,9999-12-31 23:59:59.999999]，[-15:59~+15:59]，支持的格式符：'YYYY'/'YYY'/'YY'/'Y'/'MM'/'MON'/'MONTH'/'DD'/'D'/'DAY'/'HH'/'HH12'/'HH24'/'MI'/'SS'/'AM'/'A.M.'/'PM'/'P.M.'/'FF'/'FF1'/'FF2'/'FF3'/'FF4'/'FF5'/'FF6'/'FF7'/'FF8'/'FF9'/'TZH'/'TZM'

doc: https://cod-doc.yasdb.com/yashandb/23.4/zh/Reference-Manual/Configuration-Parameters.html#timestamp-format
*/
func FormatTime(format string, t time.Time) string {
	// 替换各种格式标记
	format = strings.ToUpper(format)

	// 年
	format = strings.Replace(format, "YYYY", "2006", -1)
	format = strings.Replace(format, "YYY", "006", -1)
	format = strings.Replace(format, "YY", "06", -1)
	format = strings.Replace(format, "Y", "6", -1)

	// 月
	format = strings.Replace(format, "MONTH", "January", -1)
	format = strings.Replace(format, "MON", "Jan", -1)
	format = strings.Replace(format, "MM", "01", -1)

	// 日
	format = strings.Replace(format, "DD", "02", -1)
	format = strings.Replace(format, "D", "2", -1)
	format = strings.Replace(format, "DAY", "Monday", -1)

	// 时
	format = strings.Replace(format, "HH24", "15", -1)
	format = strings.Replace(format, "HH12", "03", -1)
	format = strings.Replace(format, "HH", "15", -1) // 默认使用24小时制

	// 分
	format = strings.Replace(format, "MI", "04", -1)
	format = strings.Replace(format, "SS", "05", -1)

	// 处理小数秒部分
	format = replaceFractionalSeconds(format)

	// 处理时区偏移
	format = replaceTimeZoneOffset(format)

	// 处理AM/PM

	format = strings.Replace(format, "AM", "PM", -1)
	toADotM := false
	if strings.Contains(format, "A.M.") {
		toADotM = true
		format = strings.Replace(format, "A.M.", "PM", -1)
	}
	format = strings.Replace(format, "PM", "PM", -1)
	toPDotM := false
	if strings.Contains(format, "A.M.") {
		toPDotM = true
		format = strings.Replace(format, "P.M.", "PM", -1)
	}

	res := t.Format(format)
	if toADotM || toPDotM {
		if isPM(t) {
			res = strings.ReplaceAll(res, "PM", "P.M.")
		} else {
			res = strings.ReplaceAll(res, "AM", "A.M.")
		}
	}
	return res
}

func isPM(t time.Time) bool {
	return t.Hour() >= 12
}

// replaceFractionalSeconds 处理小数秒部分
func replaceFractionalSeconds(format string) string {

	// 替换FF1-FF9
	for i := 1; i <= 9; i++ {
		ffTag := fmt.Sprintf("FF%d", i)
		if strings.Contains(format, ffTag) {
			// 计算指定精度的值
			frac := strings.Repeat("0", i)
			format = strings.Replace(format, ffTag, frac, -1)
		}
	}

	// 替换FF (默认6位，微秒)
	if strings.Contains(format, "FF") {
		frac := strings.Repeat("0", 6)
		format = strings.Replace(format, "FF", frac, -1)
	}

	return format
}

// replaceTimeZoneOffset 处理时区偏移
func replaceTimeZoneOffset(format string) string {
	// 处理TZH (时区小时偏移)
	if strings.Contains(format, "TZH") {
		format = strings.ReplaceAll(format, "TZH", "-07")
	}
	// 处理TZM (时区分钟偏移)
	if strings.Contains(format, "TZM") {
		format = strings.ReplaceAll(format, "TZM", "00")
	}
	return format
}

func FormatYMInterval(format string, year, month int32) string {
	var builder strings.Builder

	format = strings.ToUpper(format)

	runeFormat := []rune(format)
	for i := 0; i < len(runeFormat); {
		if runeFormat[i] != 'Y' && runeFormat[i] != 'M' {
			builder.WriteRune(runeFormat[i])
			i++
			continue
		}
		switch runeFormat[i] {
		case 'Y':
			count := 1
			for j := i + 1; j < len(runeFormat); j++ {
				if format[j] == 'Y' {
					count++
				} else {
					break
				}
			}
			if year < 0 {
				builder.WriteRune('-')
				year = -year
			}
			strYear := strconv.Itoa(int(year))
			prefixZero := count - len(strYear)
			for prefixZero > 0 {
				builder.WriteRune('0')
				prefixZero--
			}
			builder.WriteString(strYear)
			i += count
		default:
			count := 1
			if i+1 < len(runeFormat) && runeFormat[i+1] == 'M' {
				count = 2
			}
			strMonth := strconv.Itoa(int(month))
			if count == 2 {
				strMonth = fmt.Sprintf("%02d", month)
			}
			builder.WriteString(strMonth)
			i += count
		}
	}
	return builder.String()
}

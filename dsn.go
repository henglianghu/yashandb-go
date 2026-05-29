/*
Copyright  2022, YashanDB and/or its affiliates. All rights reserved.
YashanDB Driver for golang is licensed under the terms of the mulan PSL v2.0

License: 	http://license.coscl.org.cn/MulanPSL2
Home page: 	https://www.yashandb.com/
*/

package yasdb

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	dsnRegExpr       = `^(.*?)/(.*?)@(.*?)(\?(.*?))?$`
	ipv4UrlRegExpr   = `^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d+$`
	ipv6UrlRegExpr   = `^\[[:|\d|A-Z|a-z|%]+\]:\d+$`
	mappedUrlRegExpr = `^\[[:|\d|A-Z|a-z|\.]+\]:\d+$`
	udsRegExpr       = `^(.*?)(\?(.*?))?$`

	_UkeyName        = `ukey_name`
	_UkeyPin         = `ukey_pin`
	_Autocommit      = "autocommit"
	_HeartbeatEnable = "heartbeat_enable"
	_NumberAsString  = "number_as_string"
	_CompatVector    = "compat_vector"
	_CliPrepare      = "cli_prepare"
	_SslRootCer      = "ssl_root_cer"

	_DbTimestampFormat   = "timestamp_format"
	_DbDateFormat        = "date_format"
	_DbTimeFormat        = "time_format"
	_DbTimestampTzFormat = "timestamp_tz_format"
	_DbDsIntervalFormat  = "ds_interval_format"
	_DbYmIntervalFormat  = "ym_interval_format"
)

var (
	specialChars = map[string]string{
		`\`: "{dsn_placeholder_1}",
		`/`: "{dsn_placeholder_2}",
		`@`: "{dsn_placeholder_3}",
	}
	escapeChar = `\`
)

type CompatVector string

const (
	CV_NULL   CompatVector = "null"
	CV_YASHAN CompatVector = "yashan"
	CV_MYSQL  CompatVector = "mysql"
)

type DataSourceName struct {
	User              string
	Password          string
	Url               string
	DataPath          string
	IsAutoCommit      bool
	ukeyName          string
	ukeyPin           string
	heartbeatEnable   bool
	numberAsString    bool
	compatVector      CompatVector
	cliPrepare        bool
	timestampFormat   string
	timestampTzFormat string
	dateFormat        string
	timeFormat        string
	dsIntervalFormat  string
	ymIntervalFormat  string
	sslRootCer        string
}

// ParseDSN parses a DataSourceName used to connect to YashanDB
//
// It expects to receive a string in the form:
//
// [username/[password]@]host[:port][?param1=value1&...&paramN=valueN]
// OR
// YASDB_DATA_PATH[?param1=value1&...&paramN=valueN]
// Supported parameters are:
//
// autocommit - When it is true, the transaction will be automatically committed every time an SQL statement is executed. Default is false
func ParseDSN(dsnStr string) (*DataSourceName, error) {
	if dsnStr == "" {
		return nil, ErrDsnNoSet()
	}
	if isDsn(dsnStr) {
		return parseDSN(dsnStr)
	}
	return parseUDS(dsnStr)
}

func parseDSN(dsnStr string) (*DataSourceName, error) {
	dsnStr = replaceSpecialChars(dsnStr)
	dsnReg, _ := regexp.Compile(dsnRegExpr)

	if !dsnReg.MatchString(dsnStr) {
		return nil, ErrDsnNoStandard(dsnStr)
	}
	matchStrs := dsnReg.FindStringSubmatch(dsnStr)
	dsn := &DataSourceName{
		User:       recoverySpecialChars(matchStrs[1]),
		Password:   recoverySpecialChars(matchStrs[2]),
		Url:        matchStrs[3],
		DataPath:   "",
		cliPrepare: true,
	}

	// if err := checkUrl(dsn.Url); err != nil {
	// 	return nil, err
	// }

	if err := parseParams(dsn, matchStrs[4]); err != nil {
		return nil, err
	}
	genUkeyUrl(dsn)
	fillFormat(dsn)
	return dsn, nil
}

func parseUDS(dsnStr string) (*DataSourceName, error) {
	udsReg, _ := regexp.Compile(udsRegExpr)
	if !udsReg.MatchString(dsnStr) {
		return nil, ErrDsnNoStandard(dsnStr)
	}
	matchStrs := udsReg.FindStringSubmatch(dsnStr)
	dsn := &DataSourceName{
		User:       "sys",
		Password:   "",
		Url:        "",
		DataPath:   matchStrs[1],
		cliPrepare: true,
	}
	_, err := os.Stat(dsn.DataPath)
	if err != nil && !os.IsExist(err) {
		return nil, ErrDataPathNoExist(dsnStr)
	}
	if err := parseParams(dsn, matchStrs[2]); err != nil {
		return nil, err
	}
	fillFormat(dsn)
	return dsn, nil
}

func fillFormat(dsn *DataSourceName) {
	dsn.dsIntervalFormat = getOrDefault(dsn.dsIntervalFormat, _DefaultDbDsIntervalFormat)
	dsn.ymIntervalFormat = getOrDefault(dsn.ymIntervalFormat, _DefaultDbYmIntervalFormat)
	dsn.dateFormat = getOrDefault(dsn.dateFormat, _DefaultDbDateFormat)
	dsn.timeFormat = getOrDefault(dsn.timeFormat, _DefaultDbTimeFormat)
	dsn.timestampFormat = getOrDefault(dsn.timestampFormat, _DefaultDbTimestampFormat)
	dsn.timestampTzFormat = getOrDefault(dsn.timestampTzFormat, _DefaultDbTimestampTzFormat)
}

func parseParams(dsn *DataSourceName, argStr string) error {
	if argStr == "" {
		return nil
	}
	paramStr := argStr
	if argStr[0] == '?' || argStr[0] == '&' {
		paramStr = argStr[1:]
	}
	connParams := strings.Split(paramStr, "&")
	for _, param := range connParams {
		strs := strings.Split(param, "=")
		if len(strs) < 2 {
			return ErrDsnNoStandard(argStr)
		}
		switch strings.ToLower(strs[0]) {
		case _Autocommit:
			value := strings.ToLower(strs[1])
			if value == "1" || value == "true" {
				dsn.IsAutoCommit = true
			}
		case _UkeyName:
			dsn.ukeyName = strs[1]
		case _UkeyPin:
			dsn.ukeyPin = strs[1]
		case _HeartbeatEnable:
			value := strings.ToLower(strs[1])
			if value == "1" || value == "true" {
				dsn.heartbeatEnable = true
			}
		case _NumberAsString:
			value := strings.ToLower(strs[1])
			if value == "1" || value == "true" {
				dsn.numberAsString = true
			}
		case _CompatVector:
			value := strings.ToLower(strs[1])
			switch CompatVector(value) {
			case CV_MYSQL, CV_YASHAN, CV_NULL:
				dsn.compatVector = CompatVector(value)
			default:
				return fmt.Errorf("unknow compat_vector %s", value)
			}
		case _CliPrepare, "cliprepare": // 兼容旧版本
			value := strings.ToLower(strs[1])
			if value == "0" || value == "false" {
				dsn.cliPrepare = false
			}
		case _SslRootCer:
			dsn.sslRootCer = strs[1]
		case _DbDateFormat:
			dsn.dateFormat = strs[1]
		case _DbTimeFormat:
			dsn.timeFormat = strs[1]
		case _DbTimestampFormat:
			dsn.timestampFormat = strs[1]
		case _DbTimestampTzFormat:
			dsn.timestampTzFormat = strs[1]
		case _DbDsIntervalFormat:
			dsn.dsIntervalFormat = strs[1]
		case _DbYmIntervalFormat:
			dsn.ymIntervalFormat = strs[1]
		default:
			return fmt.Errorf("unknown param %s", strs[0])
		}
	}

	return nil
}

func getOrDefault(value, defaultValue string) string {
	if len(value) != 0 {
		return value
	}
	return defaultValue
}

func genUkeyUrl(dsn *DataSourceName) {
	if dsn.ukeyName != "" && dsn.ukeyPin != "" {
		dsn.Url += fmt.Sprintf("?%s=%s&%s=%s", _UkeyName, dsn.ukeyName, _UkeyPin, dsn.ukeyPin)
	} else if dsn.ukeyName != "" {
		dsn.Url += fmt.Sprintf("?%s=%s", _UkeyName, dsn.ukeyName)
	} else if dsn.ukeyPin != "" {
		dsn.Url += fmt.Sprintf("?%s=%s", _UkeyPin, dsn.ukeyPin)
	}
}

func isDsn(dsnStr string) bool {
	dsnReg, _ := regexp.Compile(dsnRegExpr)
	return dsnReg.MatchString(dsnStr)
}

func replaceSpecialChars(dsnStr string) string {
	for k, v := range specialChars {
		dsnStr = strings.ReplaceAll(dsnStr, escapeChar+k, v)
	}
	return dsnStr
}

func recoverySpecialChars(str string) string {
	for k, v := range specialChars {
		str = strings.ReplaceAll(str, v, k)
	}
	return str
}

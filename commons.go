/*
Copyright  2022, YashanDB and/or its affiliates. All rights reserved.
YashanDB Driver for golang is licensed under the terms of the mulan PSL v2.0

License: 	http://license.coscl.org.cn/MulanPSL2
Home page: 	https://www.yashandb.com/
*/

package yasdb

/*
#cgo CFLAGS: -I./yacapi/include -I./yacapi/src
#cgo !windows LDFLAGS: -ldl

#include "yacapi.h"
#include "yapi_inc.h"
#include "yacapi.go.h"
#include <stdio.h>
#include <stdlib.h>

*/
import "C"

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"
	"unsafe"
)

const (
	_LobBufLen      = 8192
	_OutputBindSize = 65535
	_DefaultSize    = 32*1024 + 1
)

type valueFreeType int8
type cliSqlType int8

var (
	CST_UNKNOW cliSqlType = 0
	CST_SELECT cliSqlType = 1
	CST_INSERT cliSqlType = 2
	CST_UPDATE cliSqlType = 3
	CST_DELETE cliSqlType = 4
	CST_PLSQL  cliSqlType = 126
)
var (
	mutex = &sync.Mutex{}

	yapiErr C.YapiErrorInfo

	byteBufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, _LobBufLen)
		},
	}

	keySqls = []string{
		"create or replace procedure",
		"create procedure",
		"create or replace trigger",
		"create trigger",
		"create or replace function",
		"create function",
		"create or replace package",
		"create or replace editionable package",
		"create package",
		"create or replace package body",
		"create package body",
		"create or replace type body",
		"create type body",
		"create or replace library",
		"begin",
		"declare",
	}

	notFree    valueFreeType = 0
	normalFree valueFreeType = 1
	lobFree    valueFreeType = 2
	cursorFree valueFreeType = 3
	vectorFree valueFreeType = 4

	commentRegStr_1 = `\/\*([^*]|\*+[^*/])*\*+\/`
	commentReg1, _  = regexp.Compile(commentRegStr_1)
	commentRegStr_2 = `^--.*`
	commentReg2, _  = regexp.Compile(commentRegStr_2)
)

type bindStruct struct {
	direction C.YapiParamDirection
	yacType   C.YapiType
	value     C.YapiPointer
	bindSize  C.int32_t
	bufLength C.int32_t
	indicator *C.int32_t
	out       sql.Out
	freeType  valueFreeType
}

func stringToYasChar(str string) *C.char {
	p := C.malloc(C.size_t(len(str) + 1))
	pp := (*[1 << 30]byte)(p)
	copy(pp[:], str)
	pp[len(str)] = 0
	return (*C.char)(p)
}

func stringToYasCharBySize(size C.size_t) *C.char {
	p := C.malloc(size + 1)
	pp := (*[1 << 30]byte)(p)
	for i := 0; i <= int(size); i++ {
		pp[i] = 0
	}
	return (*C.char)(p)
}

func intToYacInt16(n int) C.int16_t {
	return C.int16_t(n)
}

func intToYacUint16(n int) C.uint16_t {
	return C.uint16_t(n)
}

func intToYacInt32(n int) C.int32_t {
	return C.int32_t(n)
}

func intToYacUint32(n int) C.uint32_t {
	return C.uint32_t(n)
}

func intToYacInt(n int) C.int {
	return C.int(n)
}

func yacPointerToInt64(p C.YapiPointer) int64 {
	return int64(*(*C.int64_t)(p))
}

func yacPointerToInt32(p C.YapiPointer) int32 {
	return int32(*(*C.int32_t)(p))
}

func yacPointerToInt16(p C.YapiPointer) int16 {
	return int16(*(*C.int16_t)(p))
}

func yacPointerToInt8(p C.YapiPointer) int8 {
	return int8(*(*C.int8_t)(p))
}

func yacPointerToUint64(p C.YapiPointer) uint64 {
	return uint64(*(*C.uint64_t)(p))
}

func yacPointerToUint32(p C.YapiPointer) uint32 {
	return uint32(*(*C.uint32_t)(p))
}

func yacPointerToUint16(p C.YapiPointer) uint16 {
	return uint16(*(*C.uint16_t)(p))
}

func yacPointerToUint8(p C.YapiPointer) uint8 {
	return uint8(*(*C.uint8_t)(p))
}

func yacPointerToFloat64(p C.YapiPointer) float64 {
	return float64(*(*C.double)(p))
}

func yacPointerToFloat32(p C.YapiPointer) float32 {
	return float32(*(*C.float)(p))
}

func yacPointerToBool(p C.YapiPointer) bool {
	return bool(*(*C.bool)(p))
}

func mallocBytes(size uint32) unsafe.Pointer {
	p := C.malloc(C.size_t(size))
	pp := (*[1 << 30]byte)(p)
	return unsafe.Pointer(pp)
}

func sizeToAlign4(size uint32) uint32 {
	margin := uint32(size % 4)
	if margin == 0 {
		return size
	}
	return size + (4 - margin)
}

func freeFetchRows(rows []*yasRow) {
	if len(rows) == 0 || rows == nil {
		return
	}
	for i := 0; i < len(rows); i++ {
		if rows[i] == nil {
			continue
		}
		freeFetchRow(rows[i])
	}
}

func freeFetchRow(row *yasRow) {
	if row == nil {
		return
	}
	switch row.freeType {
	case lobFree:
		lobLocator := (**C.YapiLobLocator)(unsafe.Pointer(row.Data))
		C.yapiLobDescFree(unsafe.Pointer(*lobLocator), row.yacType)
	case normalFree:
		C.free(row.Data)
	case cursorFree:
		handle := (*C.YacHandle)(unsafe.Pointer(row.Data))
		_ = yapiReleaseHandle(handle)
	case vectorFree:
		if row.Data != nil && row.env != nil {
			// 获取实际的 descriptor 指针
			// row.Data 指向的是一个 Go 分配的 *unsafe.Pointer 变量，该变量包含实际的 descriptor
			ptrToDesc := (*unsafe.Pointer)(row.Data)
			actualDesc := *ptrToDesc
			if actualDesc != nil {
				// 使用包装函数，避免传递 Go 指针地址给 C
				yapiDescFree2(row.env, actualDesc, C.YAPI_TYPE_VECTOR)
			}
		}
	}

	if row.Indicator != nil {
		C.free(unsafe.Pointer(row.Indicator))
	}
	row.Data = nil
	row.Indicator = nil
}

func checkYasError(ret C.YapiResult) error {
	if int(ret) == 0 {
		return nil
	}
	mutex.Lock()
	defer func() {
		yapiErr.errCode = -1
		yapiErr.pos.line = 0
		yapiErr.pos.column = 0
		yapiErr.message = nil
		yapiErr.sqlState = nil
		mutex.Unlock()
	}()

	C.yapiGetLastError(&yapiErr)
	err := &YasDBError{
		Code:     int(yapiErr.errCode),
		Msg:      C.GoString(yapiErr.message),
		SqlState: C.GoString(yapiErr.sqlState),
		Line:     int(yapiErr.pos.line),
		Column:   int(yapiErr.pos.column),
	}
	if _Mode == _DebugMode && yapiErr.errCode == 0 {
		panic(errors.New("yasdb return code is zero"))
	}
	return err
}

func existYasError(ret C.YapiResult) bool {
	return int(ret) != 0
}

func tryRmSemicolon(query string) (string, cliSqlType) {
	cst := whichKeySql(query)
	if cst == CST_PLSQL {
		return query, cst
	}
	return strings.TrimSuffix(strings.TrimSpace(query), ";"), cst
}

func whichKeySql(query string) cliSqlType {
	query = rmComment(query)
	strs := strings.Fields(strings.TrimSpace(query))
	sqlStr := strings.ToLower(strings.Join(strs, " "))
	if strings.HasPrefix(sqlStr, "select ") {
		return CST_SELECT
	}
	if strings.HasPrefix(sqlStr, "insert into") {
		return CST_INSERT
	}
	if strings.HasPrefix(sqlStr, "update ") {
		return CST_UPDATE
	}
	if strings.HasPrefix(sqlStr, "delete from") {
		return CST_DELETE
	}
	for _, v := range keySqls {
		if strings.HasPrefix(sqlStr, v) {
			return CST_PLSQL
		}
	}
	return CST_UNKNOW
}

func rmComment(query string) string {
	if commentReg1.MatchString(query) {
		query = commentReg1.ReplaceAllString(query, "")
	}

	nQuery := ""
	for _, line := range strings.Split(query, "\n") {
		nline := strings.TrimSpace(line)
		if nline == "" {
			continue
		}
		if commentReg2.MatchString(nline) {
			continue
		}
		nQuery += nline + "\n"
	}
	return nQuery
}

func releaseConn(yasConn *C.YapiConnect) error {
	if yasConn == nil {
		return nil
	}
	if err := yapiReleaseConn(yasConn); err != nil {
		return err
	}
	yasConn = nil
	return nil
}

func releaseEnv(env *C.YapiEnv) error {
	if env == nil {
		return nil
	}
	if err := yapiReleaseEnv(env); err != nil {
		return err
	}
	env = nil
	return nil
}

func ConvertToNameValue(args ...any) ([]driver.NamedValue, error) {
	nargs := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		var (
			nargValue driver.Value
			err       error
		)
		v, isName := arg.(sql.NamedArg)
		if isName {
			nargs[i].Name = v.Name
			arg = v.Value
		}
		outValue, isOut := arg.(sql.Out)
		if isOut {
			nargValue = outValue
		} else {
			nargValue, err = driver.DefaultParameterConverter.ConvertValue(arg)
			if err != nil {
				return nil, err
			}
		}
		nargs[i].Ordinal = i + 1
		nargs[i].Value = nargValue
	}
	return nargs, nil
}

func GetDatabaseTypeName(yapiType uint32) string {
	switch C.YapiType(yapiType) {
	case C.YAPI_TYPE_BOOL:
		return "BOOLEAN"
	case C.YAPI_TYPE_TINYINT:
		return "TINYINT"
	case C.YAPI_TYPE_SMALLINT:
		return "SMALLINT"
	case C.YAPI_TYPE_INTEGER:
		return "INTEGER"
	case C.YAPI_TYPE_BIGINT:
		return "BIGINT"
	case C.YAPI_TYPE_UTINYINT:
		return "UTINYINT"
	case C.YAPI_TYPE_USMALLINT:
		return "USMALLINT"
	case C.YAPI_TYPE_UINTEGER:
		return "UINTEGER"
	case C.YAPI_TYPE_UBIGINT:
		return "UBIGINT"
	case C.YAPI_TYPE_FLOAT, C.YAPI_TYPE_NUMBER_FLOAT:
		return "FLOAT"
	case C.YAPI_TYPE_DOUBLE:
		return "DOUBLE"
	case C.YAPI_TYPE_NUMBER:
		return "NUMBER"
	case C.YAPI_TYPE_DATE:
		return "DATE"
	case C.YAPI_TYPE_SHORTTIME:
		return "TIME"
	case C.YAPI_TYPE_TIMESTAMP:
		return "TIMESTAMP"
	case C.YAPI_TYPE_CHAR:
		return "CHAR"
	case C.YAPI_TYPE_NCHAR:
		return "NCHAR"
	case C.YAPI_TYPE_VARCHAR:
		return "VARCHAR"
	case C.YAPI_TYPE_NVARCHAR:
		return "NVARCHAR"
	case C.YAPI_TYPE_CLOB:
		return "CLOB"
	case C.YAPI_TYPE_BLOB:
		return "BLOB"
	case C.YAPI_TYPE_BINARY:
		return "RAW"
	case C.YAPI_TYPE_ROWID:
		return "ROWID"
	case C.YAPI_TYPE_BIT:
		return "BIT"
	case C.YAPI_TYPE_NCLOB:
		return "NCLOB"
	case C.YAPI_TYPE_JSON:
		return "JSON"
	case C.YAPI_TYPE_YM_INTERVAL:
		return "INTERVAL YEAR TO MONTH"
	case C.YAPI_TYPE_DS_INTERVAL:
		return "INTERVAL DAY TO SECOND"
	case C.YAPI_TYPE_XML:
		return "XMLTYPE"
	case C.YAPI_TYPE_VECTOR:
		return "VECTOR"
	case C.YAPI_TYPE_TIMESTAMP_LTZ:
		return "TIMESTAMP WITH LOCAL TIME ZONE"
	case C.YAPI_TYPE_TIMESTAMP_TZ:
		return "TIMESTAMP WITH TIME ZONE"
	case C.YAPI_TYPE_BFILE:
		return "BFILE"
	case C.YAPI_TYPE_CURSOR:
		return "CURSOR"
	default:
		return ""
	}
}

func GetDatabaseTypeSize(yType C.YapiType) int32 {
	switch yType {
	case C.YAPI_TYPE_BOOL, C.YAPI_TYPE_TINYINT, C.YAPI_TYPE_UTINYINT:
		return 1
	case C.YAPI_TYPE_SMALLINT, C.YAPI_TYPE_USMALLINT:
		return 2
	case C.YAPI_TYPE_INTEGER, C.YAPI_TYPE_UINTEGER, C.YAPI_TYPE_FLOAT:
		return 4
	case C.YAPI_TYPE_BIGINT, C.YAPI_TYPE_DOUBLE, C.YAPI_TYPE_UBIGINT:
		return 8
	case C.YAPI_TYPE_NUMBER:
		return 22
	case C.YAPI_TYPE_DATE, C.YAPI_TYPE_SHORTDATE, C.YAPI_TYPE_SHORTTIME, C.YAPI_TYPE_TIMESTAMP, C.YAPI_TYPE_TIMESTAMP_TZ, C.YAPI_TYPE_TIMESTAMP_LTZ, C.YAPI_TYPE_YM_INTERVAL, C.YAPI_TYPE_DS_INTERVAL:
		return 12
	case C.YAPI_TYPE_CHAR, C.YAPI_TYPE_NCHAR, C.YAPI_TYPE_VARCHAR, C.YAPI_TYPE_NVARCHAR, C.YAPI_TYPE_BINARY, C.YAPI_TYPE_CLOB, C.YAPI_TYPE_BLOB, C.YAPI_TYPE_BIT, C.YAPI_TYPE_XML:
		return -1
	case C.YAPI_TYPE_ROWID:
		return 44
	default:
		return _DefaultSize
	}
}

func boolOutBindParam(dest *bool, in bool) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t) {
	bindSize = C.int32_t(unsafe.Sizeof(dest))
	p := C.malloc(C.size_t(bindSize))
	if in {
		*(*C.bool)(p) = C.bool(*dest)
	}
	value = C.YapiPointer(p)
	indicator = bindSize
	return
}

func bitOutBindParam(dest *[]byte, in bool) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t) {
	bindSize = 8
	bufLen = 8
	p := C.malloc(C.size_t(bindSize))
	if in {
		pp := (*[1 << 30]byte)(p)
		copy(pp[:], *dest)
	}
	value = C.YapiPointer(p)
	indicator = C.YAPI_NULL_DATA
	if len(*dest) != 0 {
		indicator = C.int32_t(len(*dest))
	}
	return
}

func int64OutBindParam(dest *int64, in bool) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t) {
	bindSize = C.int32_t(unsafe.Sizeof(dest))
	p := C.malloc(C.size_t(bindSize))
	if in {
		*(*C.int64_t)(p) = C.int64_t(*dest)
	}
	value = C.YapiPointer(p)
	indicator = bindSize
	return
}

func int32OutBindParam(dest *int32, in bool) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t) {
	bindSize = C.int32_t(unsafe.Sizeof(dest))
	p := C.malloc(C.size_t(bindSize))
	if in {
		*(*C.int32_t)(p) = C.int32_t(*dest)
	}
	value = C.YapiPointer(p)
	indicator = bindSize
	return
}

func int16OutBindParam(dest *int16, in bool) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t) {
	bindSize = C.int32_t(unsafe.Sizeof(dest))
	p := C.malloc(C.size_t(bindSize))
	if in {
		*(*C.int16_t)(p) = C.int16_t(*dest)
	}
	value = C.YapiPointer(p)
	indicator = bindSize
	return
}

func int8OutBindParam(dest *int8, in bool) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t) {
	bindSize = C.int32_t(unsafe.Sizeof(dest))
	p := C.malloc(C.size_t(bindSize))
	if in {
		*(*C.int8_t)(p) = C.int8_t(*dest)
	}
	value = C.YapiPointer(p)
	indicator = bindSize
	return
}

func uint64OutBindParam(dest *uint64, in bool) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t) {
	bindSize = C.int32_t(unsafe.Sizeof(dest))
	p := C.malloc(C.size_t(bindSize))
	if in {
		*(*C.uint64_t)(p) = C.uint64_t(*dest)
	}
	value = C.YapiPointer(p)
	indicator = bindSize
	return
}

func uint32OutBindParam(dest *uint32, in bool) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t) {
	bindSize = C.int32_t(unsafe.Sizeof(dest))
	p := C.malloc(C.size_t(bindSize))
	if in {
		*(*C.uint32_t)(p) = C.uint32_t(*dest)
	}
	value = C.YapiPointer(p)
	indicator = bindSize
	return
}

func uint16OutBindParam(dest *uint16, in bool) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t) {
	bindSize = C.int32_t(unsafe.Sizeof(dest))
	p := C.malloc(C.size_t(bindSize))
	if in {
		*(*C.uint16_t)(p) = C.uint16_t(*dest)
	}
	value = C.YapiPointer(p)
	indicator = bindSize
	return
}

func uint8OutBindParam(dest *uint8, in bool) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t) {
	bindSize = C.int32_t(unsafe.Sizeof(dest))
	p := C.malloc(C.size_t(bindSize))
	if in {
		*(*C.uint8_t)(p) = C.uint8_t(*dest)
	}
	value = C.YapiPointer(p)
	indicator = bindSize
	return
}

func dateOutBindParam(dest *time.Time, in bool) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t) {
	bindSize = 8
	p := C.malloc(C.size_t(bindSize))
	if in {
		*(*C.int64_t)(p) = C.int64_t(dest.UnixMicro())
	}
	value = C.YapiPointer(p)
	indicator = bindSize
	return
}

func timestampOutBindParam(dest *time.Time, _, in bool) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t, err error) {
	var timestamp C.YapiTimestamp

	bindSize = C.int32_t(unsafe.Sizeof(timestamp))
	bufLen = bindSize
	p := C.malloc(C.size_t(bindSize))
	if in {
		tpointer := (*C.YapiTimestamp)(p)
		year := C.int16_t(dest.Year())
		month := C.uint8_t(dest.Month())
		day := C.uint8_t(dest.Day())
		hour := C.uint8_t(dest.Hour())
		mintue := C.uint8_t(dest.Minute())
		second := C.uint8_t(dest.Second())
		fraction := nanoToMicro(dest.Nanosecond())
		err = yapiTimestampSetTimestamp(tpointer, year, month, day, hour, mintue, second, fraction)
		if err != nil {
			C.free(p)
			return
		}
		_, offset := dest.Zone()
		tpointer.bias = C.int16_t(offset / 60)
	}

	value = C.YapiPointer(p)
	indicator = bindSize
	return
}

func float64OutBindParam(dest *float64, in bool) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t) {
	bindSize = C.int32_t(unsafe.Sizeof(dest))
	p := C.malloc(C.size_t(bindSize))
	if in {
		*(*C.double)(p) = C.double(*dest)
	}
	value = C.YapiPointer(p)
	indicator = bindSize
	return
}

func float32OutBindParam(dest *float32, in bool) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t) {
	bindSize = C.int32_t(unsafe.Sizeof(dest))
	p := C.malloc(C.size_t(bindSize))
	if in {
		*(*C.float)(p) = C.float(*dest)
	}
	value = C.YapiPointer(p)
	indicator = bindSize
	return
}

func stringOutBindParam(dest *string, size int, in bool) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t) {
	n := len(*dest)
	bindSize = getMallocSize(n, size)
	bufLen = bindSize
	p := C.malloc(C.size_t(bindSize))
	if in {
		pp := (*[1 << 30]byte)(p)
		copy(pp[:], *dest)
		pp[n] = 0 // 添加终结符
	}

	value = C.YapiPointer(p)

	indicator = C.YAPI_NULL_DATA
	if n > 0 {
		// 需要把\0也算进去
		indicator = C.int32_t(n)
	}
	return
}

func rawOutBindParam(dest *[]byte, size int, in bool) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t) {
	n := len(*dest)
	bindSize = getMallocSize(n, size)
	p := C.malloc(C.size_t(bindSize))
	bufLen = bindSize
	if in {
		pp := (*[1 << 30]byte)(p)
		copy(pp[:], *dest)
	}
	value = C.YapiPointer(p)
	indicator = C.YAPI_NULL_DATA
	if n > 0 {
		indicator = C.int32_t(n)
	}
	return
}

func cursorOutBindParam(dest *YasRows) (bindSize, bufLen C.int32_t, value C.YapiPointer, indicator C.int32_t) {
	var handle C.YacHandle
	bufLen = C.int32_t(unsafe.Sizeof(handle))
	bindSize = bufLen
	// data := (*C.YacHandle)(C.malloc(C.size_t(bufLen)))
	data := (*C.YacHandle)(C.malloc(C.size_t(bufLen)))
	value = C.YapiPointer(data)
	return
}

func getMallocSize(actual, want int) C.int32_t {
	if want > actual {
		return C.int32_t(want)
	}
	return C.int32_t(actual)
}

func fixTimeString(s string) string {
	// fix: 2016-01-02 15:04:05. -07:00 -> 2016-01-02 15:04:05 -07:00
	s = strings.ReplaceAll(s, ". ", " ")
	// fix: 2016-01-02 15:04:05. -> 2016-01-02 15:04:05
	s = strings.TrimRight(s, ".")
	return s
}

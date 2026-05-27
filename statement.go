/*
Copyright  2022, YashanDB and/or its affiliates. All rights reserved.
YashanDB Driver for golang is licensed under the terms of the mulan PSL v2.0

License: 	http://license.coscl.org.cn/MulanPSL2
Home page: 	https://www.yashandb.com/
*/

package yasdb

/*
#cgo CFLAGS: -I./yacapi/include

#include "yacapi.h"
#include <stdio.h>
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"sync"
	"time"
	"unsafe"
)

type YasStmt struct {
	Conn     *YasConn
	Stmt     *C.YapiStmt
	closed   bool
	SqlType  uint32
	Sqlstr   string
	ctx      context.Context
	binds    []*bindStruct
	prepared bool
	sync.Mutex
}

// Query executes a query that may return rows, such as a SELECT.
//
// Deprecated: Drivers should implement StmtQueryContext instead (or additionally).
func (stmt *YasStmt) Query(args []driver.Value) (driver.Rows, error) {
	nargs := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		nargs[i].Ordinal = i + 1
		nargs[i].Value = arg
	}
	return stmt.QueryContext(context.Background(), nargs)
}

// QueryContext executes a query that may return rows, such as a SELECT.
//
// QueryContext must honor the context timeout and return when it is canceled.
func (stmt *YasStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	stmt.Lock()
	defer stmt.Unlock()
	stmt.ctx = ctx

	defer stmt.freeBindValues()
	if err := stmt.bindValues(args); err != nil {
		return nil, err
	}
	return stmt.query()
}

// Exec executes a query that doesn't return rows, such as an INSERT or UPDATE.
//
// Deprecated: Drivers should implement StmtExecContext instead (or additionally).
func (stmt *YasStmt) Exec(args []driver.Value) (driver.Result, error) {
	if args == nil {
		args = []driver.Value{}
	}
	nargs := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		nargs[i].Ordinal = i + 1
		nargs[i].Value = arg
	}

	return stmt.ExecContext(context.Background(), nargs)
}

// ExecContext executes a query that doesn't return rows, such as an INSERT or UPDATE.
//
// ExecContext must honor the context timeout and return when it is canceled.
func (stmt *YasStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	stmt.Lock()
	defer stmt.Unlock()
	stmt.ctx = ctx

	if args == nil {
		args = []driver.NamedValue{}
	}

	defer stmt.freeBindValues()
	if err := stmt.bindValues(args); err != nil {
		return nil, err
	}

	return stmt.exec()
}

// NumInput returns the number of placeholder parameters.
func (stmt *YasStmt) NumInput() int {
	if stmt.Conn == nil || stmt.Stmt == nil {
		return -1
	}
	var paramList C.YapiPointer
	sqlChar := C.CString(stmt.Sqlstr)
	defer C.free(unsafe.Pointer(sqlChar))
	err := yapiParseSqlParams(stmt.Conn.Env, &paramList, sqlChar, C.int32_t(len(stmt.Sqlstr)))
	if err != nil {
		return -1
	}
	defer yapiFreeParamList(paramList)

	var count C.uint32_t
	if err := yapiGetParamListCount(paramList, &count); err != nil {
		return -1
	}
	return int(count)
}

// Close closes the statement.
func (stmt *YasStmt) Close() error {
	if stmt.closed {
		return nil
	}
	stmt.closed = true
	return stmt.yapiReleaseStmt()
}

// CheckNamedValue is called before passing arguments to the driver
// and is called in place of any ColumnConverter. CheckNamedValue must do type
// validation and conversion as appropriate for the driver.
func (stmt *YasStmt) CheckNamedValue(namedValue *driver.NamedValue) error {
	switch namedValue.Value.(type) {
	case sql.Out, DSInterval, YMInterval, Vector, *Vector, []Vector, []*Vector:
		return nil
	}
	return driver.ErrSkip
}

func (stmt *YasStmt) query() (driver.Rows, error) {
	if stmt.ctx.Err() != nil {
		return nil, stmt.ctx.Err()
	}
	var err error
	if stmt.ctx == context.Background() {
		err = stmt.yacExecute()
	} else {
		done := make(chan struct{})
		go stmt.Conn.handleYacCancel(stmt.ctx, done)
		err = stmt.yacExecute()
		close(done)
	}
	if err != nil {
		return nil, err
	}

	fetchRows, err := stmt.getFetchRows()
	if err != nil {
		return nil, err
	}
	rows := YasRows{
		stmt:      stmt,
		fetchRows: fetchRows,
	}
	return &rows, nil
}

func (stmt *YasStmt) exec() (driver.Result, error) {
	if stmt.ctx.Err() != nil {
		return nil, stmt.ctx.Err()
	}

	var err error
	if stmt.ctx == context.Background() {
		err = stmt.yacExecute()
	} else {
		done := make(chan struct{})
		go stmt.Conn.handleYacCancel(stmt.ctx, done)
		err = stmt.yacExecute()
		close(done)
	}
	if err != nil {
		return nil, err
	}

	rowsAffected, rowsAffectedErr := stmt.getRowsAffected()
	result := YasResult{
		rowsAffected:    rowsAffected,
		rowsAffectedErr: rowsAffectedErr,
	}

	if err := stmt.getBindValueDest(); err != nil {
		return nil, err
	}
	return &result, nil
}

func (stmt *YasStmt) yacExecute() error {
	if stmt.prepared {
		return yapiExecute(stmt.Stmt)
	} else {
		return yapiDirectExecute(stmt.Stmt, stmt.Sqlstr)
	}
}

func (stmt *YasStmt) yapiReleaseStmt() error {
	if stmt.Stmt == nil {
		return nil
	}
	if err := yapiReleaseStmt(stmt.Stmt); err != nil {
		return err
	}
	stmt.Stmt = nil
	return nil
}

func (stmt *YasStmt) getFetchRows() ([]*yasRow, error) {
	columns := C.int16_t(0)
	if err := yapiNumResultCols(stmt.Stmt, &columns); err != nil {
		return nil, err
	}
	columnCount := int(columns)
	yasRows := make([]*yasRow, 0, columnCount)
	for i := 0; i < columnCount; i++ {
		row, err := stmt.getFetchRow(i)
		if err != nil {
			freeFetchRows(yasRows)
			return nil, err
		}
		yasRows = append(yasRows, row)
	}
	return yasRows, nil
}

func (stmt *YasStmt) getFetchRow(pos int) (*yasRow, error) {
	item := C.YapiColumnDesc{}
	if err := yapiDescribeCol2(stmt.Stmt, C.uint16_t(pos), &item); err != nil {
		return nil, err
	}
	yacType := C.YapiType(item._type)
	size := uint32(item.size)
	precision := yapiColumnDescGetPrecision(&item)
	scale := yapiColumnDescGetScale(&item)
	nullable := uint8(item.nullable)
	row := NewYasRow(size, yacType, precision, scale, nullable)
	row.env = stmt.Conn.Env // 添加 env 引用，用于释放 vector 资源
	bufLen := int32(size)
	freeType := notFree

	switch yacType {
	case C.YAPI_TYPE_NCHAR, C.YAPI_TYPE_NVARCHAR:
		yacType = C.YAPI_TYPE_VARCHAR
		bufLen = int32(sizeToAlign4(size)*stmt.Conn.ncharsetRatio) + 1
		row.Data = mallocBytes(uint32(bufLen))
		freeType = normalFree
	case C.YAPI_TYPE_CHAR, C.YAPI_TYPE_VARCHAR:
		bufLen = int32(sizeToAlign4(size)*stmt.Conn.charsetRatio) + 1
		row.Data = mallocBytes(uint32(bufLen))
		freeType = normalFree
	case C.YAPI_TYPE_NUMBER: // number to string
		yacType = C.YAPI_TYPE_VARCHAR
		bufLen = int32(sizeToAlign4(uint32(precision) + 8))
		row.Data = mallocBytes(uint32(bufLen))
		freeType = normalFree
	case C.YAPI_TYPE_YM_INTERVAL:
		var cymi C.YapiYMInterval
		bufLen = int32(unsafe.Sizeof(cymi))
		row.Data = mallocBytes(uint32(bufLen))
		freeType = normalFree
	case C.YAPI_TYPE_DS_INTERVAL:
		var cdsi C.YapiDSInterval
		bufLen = int32(unsafe.Sizeof(cdsi))
		row.Data = mallocBytes(uint32(bufLen))
		freeType = normalFree
	case C.YAPI_TYPE_DATE, C.YAPI_TYPE_TIMESTAMP, C.YAPI_TYPE_SHORTDATE, C.YAPI_TYPE_SHORTTIME, C.YAPI_TYPE_TIMESTAMP_LTZ:
		bufLen = 12
		row.Data = mallocBytes(uint32(bufLen))
		freeType = normalFree
	case C.YAPI_TYPE_TIMESTAMP_TZ:
		yacType = C.YAPI_TYPE_VARCHAR
		bufLen = 40
		row.Data = mallocBytes(uint32(bufLen))
		freeType = normalFree
	case C.YAPI_TYPE_CLOB, C.YAPI_TYPE_BLOB, C.YAPI_TYPE_XML, C.YAPI_TYPE_NCLOB:
		desc := new(unsafe.Pointer)
		if err := yapiLobDescAlloc(stmt.Conn.Conn, yacType, desc); err != nil {
			return nil, err
		}
		bufLen = -1
		row.Data = unsafe.Pointer(desc)
		freeType = lobFree
	case C.YAPI_TYPE_BOOL, C.YAPI_TYPE_TINYINT, C.YAPI_TYPE_SMALLINT, C.YAPI_TYPE_INTEGER, C.YAPI_TYPE_BIGINT:
		row.Data = mallocBytes(size)
		freeType = normalFree
	case C.YAPI_TYPE_UTINYINT, C.YAPI_TYPE_USMALLINT, C.YAPI_TYPE_UINTEGER, C.YAPI_TYPE_UBIGINT:
		row.Data = mallocBytes(size)
		freeType = normalFree
	case C.YAPI_TYPE_VECTOR:
		// Vector 类型使用描述符方式，需要分配 YapiVector 指针的指针
		desc := new(unsafe.Pointer)
		if err := yapiDescAlloc2(stmt.Conn.Env, desc, C.YAPI_DESC_VECTOR); err != nil {
			return nil, err
		}
		bufLen = -1
		row.Data = unsafe.Pointer(desc)
		freeType = vectorFree
	case C.YAPI_TYPE_FLOAT, C.YAPI_TYPE_DOUBLE, C.YAPI_TYPE_BIT:
		row.Data = mallocBytes(size)
		freeType = normalFree
	case C.YAPI_TYPE_BINARY:
		bufLen = int32(size*2 + 1)
		if bufLen < _DefaultSize {
			// 视图V$COLUMN_STATISTICS_CACHE的LOWVAL,HIGHVAL字段类型是RAW(8)，实际大小远大于8，因此使用默认最大的buffer来绑定数据
			//
			// select LOWVAL,HIGHVAL from V$COLUMN_STATISTICS_CACHE LIMIT 101 OFFSET 0
			bufLen = _DefaultSize
		}
		row.Data = mallocBytes(uint32(bufLen))
		freeType = normalFree
	case C.YAPI_TYPE_ROWID:
		yacType = C.YAPI_TYPE_VARCHAR
		bufLen = 44
		row.Data = mallocBytes(uint32(bufLen))
		freeType = normalFree
	case C.YAPI_TYPE_CURSOR:
		yacType = C.YAPI_TYPE_CURSOR
		var handle C.YacHandle
		bufLen = int32(unsafe.Sizeof(handle))
		data := (*C.YacHandle)(C.malloc(C.size_t(bufLen)))
		freeType = cursorFree
		row.Data = unsafe.Pointer(data)
	default:
		yacType = C.YAPI_TYPE_VARCHAR
		bufLen = _DefaultSize
		row.Data = mallocBytes(uint32(bufLen) * stmt.Conn.charsetRatio)
		freeType = normalFree
	}
	indicator := (*C.int32_t)(C.malloc(4))
	*indicator = 0
	row.Indicator = indicator
	row.freeType = freeType
	if err := yapiBindColumn(
		stmt.Stmt,
		C.uint16_t(pos),
		yacType,
		C.YapiPointer(row.Data),
		C.int32_t(bufLen),
		indicator,
	); err != nil {
		freeFetchRow(row)
		return nil, err
	}
	row.name = C.GoString(item.name)
	return row, nil
}

func (stmt *YasStmt) getRowsAffected() (int64, error) {
	var rowsCount C.uint32_t
	size := C.int32_t(unsafe.Sizeof(rowsCount))
	sLength := C.int32_t(0)
	err := yapiGetStmtAttr(
		stmt.Stmt,
		C.YAPI_ATTR_ROWS_AFFECTED,
		unsafe.Pointer(&rowsCount),
		size,
		sLength,
	)
	return int64(rowsCount), err
}

func (stmt *YasStmt) bindValues(args []driver.NamedValue) error {
	if len(args) == 0 {
		return nil
	}
	stmt.binds = make([]*bindStruct, 0, len(args))
	var err error
	for index, narg := range args {
		arg := narg.Value
		sqlOut, isOut := arg.(sql.Out)
		bind := &bindStruct{}

		if isOut {
			bind, err = stmt.getOutputBindValue(sqlOut)
			bind.out = sqlOut
		} else {
			bind, err = stmt.getInputBindValue(arg)
		}
		if err != nil {
			return err
		}

		stmt.binds = append(stmt.binds, bind)

		if len(narg.Name) == 0 {
			err = stmt.yacBindParameter(bind, intToYacUint16(index+1))
		} else {
			err = stmt.yacBindParameterByName(bind, narg.Name)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (stmt *YasStmt) yacBindParameter(b *bindStruct, pos C.uint16_t) error {
	if err := yapiBindParameter(stmt.Stmt, b, pos); err != nil {
		return err
	}
	return nil
}

func (stmt *YasStmt) yacBindParameterByName(b *bindStruct, name string) error {
	charName := stringToYasChar(name)
	defer C.free(unsafe.Pointer(charName))
	if err := yapiBindParameterByName(stmt.Stmt, charName, b); err != nil {
		return err
	}
	return nil
}

func (stmt *YasStmt) getInputBindValue(arg driver.Value) (*bindStruct, error) {
	bind := &bindStruct{}
	var (
		yacType   C.YapiType
		bindSize  C.int32_t
		value     C.YapiPointer
		indicator *C.int32_t
		bufLength C.int32_t
		freeType  valueFreeType
	)

	bindSize = C.int32_t(unsafe.Sizeof(arg)) + 1
	bufLength = C.int32_t(bindSize - 1)
	indicator = new(C.int32_t)
	*indicator = C.int32_t(bindSize - 1)
	freeType = notFree

	switch v := arg.(type) {
	case int64:
		yacType = C.YAPI_TYPE_BIGINT
		value = C.YapiPointer(unsafe.Pointer(&v))
	case float64:
		yacType = C.YAPI_TYPE_DOUBLE
		value = C.YapiPointer(unsafe.Pointer(&v))
	case bool:
		yacType = C.YAPI_TYPE_BOOL
		value = C.YapiPointer(unsafe.Pointer(&v))
	case string:
		yacType = C.YAPI_TYPE_VARCHAR
		bindSize = C.int32_t(len(v)) + 1
		bufLength = C.int32_t(bindSize - 1)
		indicator = nil
		value = C.YapiPointer(unsafe.Pointer(stringToYasChar(v)))
		freeType = normalFree
	case []byte:
		desc, err := stmt.Conn.lobWrite(C.YAPI_TYPE_BLOB, v)
		if err != nil {
			return nil, err
		}
		yacType = C.YAPI_TYPE_BLOB
		bindSize = -1
		bufLength = -1
		indicator = nil
		value = C.YapiPointer(desc)
		freeType = lobFree
	case time.Time:
		yacType = C.YAPI_TYPE_TIMESTAMP
		timestamp, err := stmt.Conn.timeToYapiTimestamp(&v)
		if err != nil {
			return nil, err
		}
		value = C.YapiPointer(unsafe.Pointer(timestamp))
		freeType = normalFree
	case DSInterval:
		yacType = C.YAPI_TYPE_DS_INTERVAL
		dsi, err := stmt.Conn.dsIntervalToYapiDSInterval(v)
		if err != nil {
			return nil, err
		}
		value = C.YapiPointer(dsi)
		freeType = normalFree
	case YMInterval:
		yacType = C.YAPI_TYPE_YM_INTERVAL
		ymi, err := stmt.Conn.ymIntervalToYapiYmInterval(v)
		if err != nil {
			return nil, err
		}
		value = C.YapiPointer(ymi)
		freeType = normalFree
	case nil:
		yacType = C.YAPI_TYPE_CHAR
		bindSize = 0
		*indicator = C.YAPI_NULL_DATA
		value = C.YapiPointer(unsafe.Pointer(&v))
	case Vector, *Vector, []Vector, []*Vector:
		info, err := bindVectorValue(stmt.Conn.Env, v)
		if err != nil {
			return nil, err
		}
		yacType = info.yacType
		bindSize = info.bindSize
		bufLength = info.bufLength
		indicator = info.indicator
		value = info.value
		freeType = info.freeType
	default:
		return nil, ErrUnknowType(arg)
	}

	bind.yacType = yacType
	bind.value = value
	bind.bindSize = bindSize
	bind.bufLength = bufLength
	bind.indicator = indicator
	bind.direction = C.YAPI_PARAM_INPUT
	bind.freeType = freeType
	return bind, nil
}

func (stmt *YasStmt) getOutputBindValue(sqlOut sql.Out) (*bindStruct, error) {
	if obi, ok := sqlOut.Dest.(*outputBindInfo); ok {
		return stmt.getOutputBindValueByInfo(obi, sqlOut.In)
	} else {
		return stmt.getOutputBindValueByDest(sqlOut.Dest, sqlOut.In)
	}
}

func (stmt *YasStmt) getOutputBindValueByDest(dest interface{}, inout bool) (*bindStruct, error) {
	bind := &bindStruct{}
	var (
		yacType   C.YapiType
		bindSize  C.int32_t
		value     C.YapiPointer
		indicator *C.int32_t
		bufLength C.int32_t
		arg       driver.Value
		err       error

		freeType = notFree
	)

	arg, err = driver.DefaultParameterConverter.ConvertValue(dest)
	if err != nil {
		return bind, err
	}

	switch arg.(type) {
	case nil:
		arg = dest
		switch arg.(type) {
		case *sql.NullBool:
			arg = false
		case *sql.NullFloat64:
			arg = float64(0)
		case *sql.NullInt64:
			arg = int64(0)
		case *sql.NullString:
			arg = ""
		}
	}

	bindSize = C.int32_t(0)
	bufLength = C.int32_t(0)
	indicator = new(C.int32_t)

	switch v := arg.(type) {
	case int64:
		yacType = C.YAPI_TYPE_BIGINT
		bindSize, bufLength, value, *indicator = int64OutBindParam(&v, inout)
		freeType = normalFree
	case float64:
		yacType = C.YAPI_TYPE_DOUBLE
		bindSize, bufLength, value, *indicator = float64OutBindParam(&v, inout)
		freeType = normalFree
	case bool:
		yacType = C.YAPI_TYPE_BOOL
		bindSize, bufLength, value, *indicator = boolOutBindParam(&v, inout)
		freeType = normalFree
	case string:
		yacType = C.YAPI_TYPE_VARCHAR
		bindSize, bufLength, value, *indicator = stringOutBindParam(&v, _OutputBindSize, inout)
		freeType = normalFree
	case []byte:
		desc, err := stmt.Conn.lobWrite(C.YAPI_TYPE_BLOB, v)
		if err != nil {
			return bind, err
		}
		yacType = C.YAPI_TYPE_BLOB
		bindSize = -1
		bufLength = -1
		indicator = nil
		value = C.YapiPointer(desc)
		freeType = lobFree
	case time.Time:
		yacType = C.YAPI_TYPE_TIMESTAMP
		bindSize, bufLength, value, *indicator, err = timestampOutBindParam(&v, false, inout)
		if err != nil {
			return bind, err
		}
		freeType = normalFree
	case nil:
		yacType = C.YAPI_TYPE_CHAR
		bindSize = 0
		*indicator = C.YAPI_NULL_DATA
		value = C.YapiPointer(unsafe.Pointer(&v))
	default:
		return bind, ErrUnknowType(v)
	}

	bind.yacType = yacType
	bind.value = value
	bind.bindSize = bindSize
	bind.bufLength = bufLength
	bind.indicator = indicator
	bind.freeType = freeType
	bind.direction = C.YAPI_PARAM_OUTPUT
	if inout {
		bind.direction = C.YAPI_PARAM_INOUT
	}
	return bind, nil
}

func (stmt *YasStmt) getOutputBindValueByInfo(obi *outputBindInfo, inout bool) (*bindStruct, error) {
	bind := &bindStruct{}
	var (
		yacType   C.YapiType = obi.yacType
		bindSize  C.int32_t
		value     C.YapiPointer
		indicator *C.int32_t
		bufLength C.int32_t

		freeType = notFree
	)

	if obi.bindSize == 0 {
		bindSize = _OutputBindSize
	} else {
		bindSize = obi.bindSize + 1
	}
	bufLength = C.int32_t(bindSize)
	indicator = new(C.int32_t)
	*indicator = C.int32_t(bindSize - 1)

	switch yacType {
	case C.YAPI_TYPE_BLOB:
		v, err := obi.getBlobBindDest()
		if err != nil {
			return bind, err
		}
		desc, err := stmt.Conn.lobWrite(C.YAPI_TYPE_BLOB, *v)
		if err != nil {
			return bind, err
		}
		bindSize = -1
		bufLength = -1
		indicator = nil
		value = C.YapiPointer(desc)
		freeType = lobFree
	case C.YAPI_TYPE_CLOB:
		v, err := obi.getClobBindDest()
		if err != nil {
			return bind, err
		}
		desc, err := stmt.Conn.lobWrite(C.YAPI_TYPE_CLOB, []byte(*v))
		if err != nil {
			return bind, err
		}
		bindSize = -1
		bufLength = -1
		indicator = nil
		value = C.YapiPointer(desc)
		freeType = lobFree
	case C.YAPI_TYPE_BIT:
		v, err := obi.getBitBindDest()
		if err != nil {
			return bind, err
		}
		bindSize, bufLength, value, *indicator = bitOutBindParam(v, inout)
		freeType = normalFree
	case C.YAPI_TYPE_BOOL:
		v, err := obi.getBoolBindDest()
		if err != nil {
			return bind, err
		}
		bindSize, bufLength, value, *indicator = boolOutBindParam(v, inout)
		freeType = normalFree
	case C.YAPI_TYPE_TINYINT:
		v, err := obi.getInt8BindDest()
		if err != nil {
			return bind, err
		}
		bindSize, bufLength, value, *indicator = int8OutBindParam(v, inout)
		freeType = normalFree
	case C.YAPI_TYPE_SMALLINT:
		v, err := obi.getInt16BindDest()
		if err != nil {
			return bind, err
		}
		bindSize, bufLength, value, *indicator = int16OutBindParam(v, inout)
	case C.YAPI_TYPE_INTEGER:
		v, err := obi.getInt32BindDest()
		if err != nil {
			return bind, err
		}
		bindSize, bufLength, value, *indicator = int32OutBindParam(v, inout)
	case C.YAPI_TYPE_BIGINT:
		v, err := obi.getInt64BindDest()
		if err != nil {
			return bind, err
		}
		bindSize, bufLength, value, *indicator = int64OutBindParam(v, inout)
		freeType = normalFree
	case C.YAPI_TYPE_UTINYINT:
		v, err := obi.getUint8BindDest()
		if err != nil {
			return bind, err
		}
		bindSize, bufLength, value, *indicator = uint8OutBindParam(v, inout)
		freeType = normalFree
	case C.YAPI_TYPE_USMALLINT:
		v, err := obi.getUint16BindDest()
		if err != nil {
			return bind, err
		}
		bindSize, bufLength, value, *indicator = uint16OutBindParam(v, inout)
	case C.YAPI_TYPE_UINTEGER:
		v, err := obi.getUint32BindDest()
		if err != nil {
			return bind, err
		}
		bindSize, bufLength, value, *indicator = uint32OutBindParam(v, inout)
	case C.YAPI_TYPE_UBIGINT:
		v, err := obi.getUint64BindDest()
		if err != nil {
			return bind, err
		}
		bindSize, bufLength, value, *indicator = uint64OutBindParam(v, inout)
		freeType = normalFree
	case C.YAPI_TYPE_DATE:
		v, err := obi.getTimeBindDest()
		if err != nil {
			return bind, err
		}
		bindSize, bufLength, value, *indicator = dateOutBindParam(v, inout)
		freeType = normalFree
	case C.YAPI_TYPE_TIMESTAMP, C.YAPI_TYPE_TIMESTAMP_LTZ, C.YAPI_TYPE_TIMESTAMP_TZ:
		v, err := obi.getTimeBindDest()
		if err != nil {
			return bind, err
		}
		zone := false
		if yacType == C.YAPI_TYPE_TIMESTAMP_TZ {
			zone = true
		}
		bindSize, bufLength, value, *indicator, err = timestampOutBindParam(v, zone, inout)
		if err != nil {
			return bind, err
		}
		freeType = normalFree
	case C.YAPI_TYPE_BINARY:
		v, err := obi.getBlobBindDest()
		if err != nil {
			return bind, err
		}
		bindSize, bufLength, value, *indicator = rawOutBindParam(v, int(bindSize), inout)
		freeType = normalFree
	case C.YAPI_TYPE_DOUBLE:
		v, err := obi.getFloat64BindDest()
		if err != nil {
			return bind, err
		}
		bindSize, bufLength, value, *indicator = float64OutBindParam(v, inout)
		freeType = normalFree
	case C.YAPI_TYPE_FLOAT:
		v, err := obi.getFloat32BindDest()
		if err != nil {
			return bind, err
		}
		bindSize, bufLength, value, *indicator = float32OutBindParam(v, inout)
		freeType = normalFree
	case C.YAPI_TYPE_DS_INTERVAL:
		v, err := obi.getIntervalBindDest()
		if err != nil {
			return bind, err
		}
		dsInterval, err := stmt.Conn.stringToYapiDSInterval(v)
		if err != nil {
			return bind, err
		}
		bindSize = C.int32_t(unsafe.Sizeof(*dsInterval))
		bufLength = bindSize
		*indicator = C.int32_t(bufLength)
		value = C.YapiPointer(dsInterval)
		freeType = normalFree
	case C.YAPI_TYPE_YM_INTERVAL:
		v, err := obi.getIntervalBindDest()
		if err != nil {
			return bind, err
		}
		ymInterval, err := stmt.Conn.stringToYapiYMInterval(v)
		if err != nil {
			return bind, err
		}
		bindSize = C.int32_t(unsafe.Sizeof(*ymInterval))
		bufLength = bindSize
		*indicator = C.int32_t(bufLength)
		value = C.YapiPointer(ymInterval)
		freeType = normalFree
	case C.YAPI_TYPE_CHAR, C.YAPI_TYPE_VARCHAR:
		yacType = C.YAPI_TYPE_VARCHAR
		v, err := obi.getVarcharBindDest()
		if err != nil {
			return bind, err
		}
		size := int(sizeToAlign4(uint32(bindSize))*stmt.Conn.charsetRatio) + 1
		bindSize, bufLength, value, *indicator = stringOutBindParam(v, size, inout)
		freeType = normalFree
	case C.YAPI_TYPE_NCHAR, C.YAPI_TYPE_NVARCHAR:
		yacType = C.YAPI_TYPE_VARCHAR
		v, err := obi.getVarcharBindDest()
		if err != nil {
			return bind, err
		}
		size := int(sizeToAlign4(uint32(bindSize))*stmt.Conn.ncharsetRatio) + 1
		bindSize, bufLength, value, *indicator = stringOutBindParam(v, size, inout)
		freeType = normalFree
	case C.YAPI_TYPE_NUMBER:
		v, err := obi.getNumberDest()
		if err != nil {
			return bind, err
		}
		number, err := stmt.Conn.float64ToYapiNumber(v)
		if err != nil {
			return bind, err
		}
		bindSize = C.int32_t(unsafe.Sizeof(*number))
		bufLength = bindSize
		*indicator = C.int32_t(bufLength)
		value = C.YapiPointer(number)
		freeType = normalFree

	case C.YAPI_TYPE_ROWID:
		yacType = C.YAPI_TYPE_VARCHAR
		v, err := obi.getVarcharBindDest()
		if err != nil {
			return bind, err
		}
		size := int(16)
		bindSize, bufLength, value, *indicator = stringOutBindParam(v, size, inout)
		freeType = normalFree

	case C.YAPI_TYPE_CURSOR:
		yacType = C.YAPI_TYPE_CURSOR
		v, err := obi.getCursorBindDest()
		if err != nil {
			return bind, err
		}
		bindSize, bufLength, value, *indicator = cursorOutBindParam(v)
		freeType = cursorFree
	case C.YAPI_TYPE_VECTOR:
		v, err := obi.getVectorBindDest()
		if err != nil {
			return bind, err
		}
		// Vector 类型使用描述符方式，需要分配 YapiVector 指针的指针
		desc := new(unsafe.Pointer)
		if err := yapiDescAlloc2(stmt.Conn.Env, desc, C.YAPI_DESC_VECTOR); err != nil {
			return bind, err
		}
		vector := (*C.YapiVector)(*desc)

		// 根据数据创建 vector
		var err2 error
		switch data := v.Data.(type) {
		case []float32:
			elemSize := getVectorElementSizeByFormat(VectorFormatFloat32)
			dim := C.uint16_t(len(data))
			array := (*C.uint8_t)(unsafe.Pointer(&data[0]))
			arrayLen := C.uint32_t(len(data) * elemSize)
			err2 = yapiVectorFromArray(vector, C.YAPI_VECTOR_FORMAT_FLOAT32, dim, array, arrayLen, 0)
		case []float64:
			elemSize := getVectorElementSizeByFormat(VectorFormatFloat64)
			dim := C.uint16_t(len(data))
			array := (*C.uint8_t)(unsafe.Pointer(&data[0]))
			arrayLen := C.uint32_t(len(data) * elemSize)
			err2 = yapiVectorFromArray(vector, C.YAPI_VECTOR_FORMAT_FLOAT64, dim, array, arrayLen, 0)
		case []int8:
			dim := C.uint16_t(len(data))
			array := (*C.uint8_t)(unsafe.Pointer(&data[0]))
			arrayLen := C.uint32_t(len(data))
			err2 = yapiVectorFromArray(vector, C.YAPI_VECTOR_FORMAT_INT8, dim, array, arrayLen, 0)
		case string:
			text := C.CString(data)
			defer C.free(unsafe.Pointer(text))
			textLen := C.uint32_t(len(data))
			err2 = yapiVectorFromText(vector, C.YAPI_VECTOR_FORMAT_FLEX, 0, text, textLen, 0)
		default:
			yapiDescFree2(stmt.Conn.Env, *desc, C.YAPI_TYPE_VECTOR)
			return bind, fmt.Errorf("unsupported Vector data type: %T", v.Data)
		}

		if err2 != nil {
			yapiDescFree2(stmt.Conn.Env, *desc, C.YAPI_TYPE_VECTOR)
			return bind, err2
		}

		bindSize = -1
		bufLength = -1
		indicator = nil
		value = C.YapiPointer(unsafe.Pointer(desc))
		freeType = vectorFree
	default:
		return bind, ErrUnknowType(yacType)
	}

	bind.yacType = yacType
	bind.value = value

	bind.bindSize = bindSize
	bind.bufLength = bufLength
	bind.indicator = indicator
	bind.freeType = freeType
	bind.direction = C.YAPI_PARAM_OUTPUT
	if inout {
		bind.direction = C.YAPI_PARAM_INOUT
	}
	return bind, nil
}

func (stmt *YasStmt) getBindValueDest() error {
	var err error
	for index, bind := range stmt.binds {
		if bind.value == nil || bind.out.Dest == nil {
			continue
		}
		switch dest := bind.out.Dest.(type) {
		case *int8:
			*dest = int8(yacPointerToInt64(bind.value))
		case *int16:
			*dest = int16(yacPointerToInt64(bind.value))
		case *int32:
			*dest = int32(yacPointerToInt64(bind.value))
		case *int:
			*dest = int(yacPointerToInt64(bind.value))
		case *int64:
			*dest = yacPointerToInt64(bind.value)
		case *uint8:
			*dest = uint8(yacPointerToUint64(bind.value))
		case *uint16:
			*dest = uint16(yacPointerToUint64(bind.value))
		case *uint32:
			*dest = uint32(yacPointerToUint64(bind.value))
		case *uint:
			*dest = uint(yacPointerToUint64(bind.value))
		case *uint64:
			*dest = yacPointerToUint64(bind.value)
		case *uintptr:
			*dest = uintptr(yacPointerToUint64(bind.value))
		case *float32:
			*dest = float32(yacPointerToFloat64(bind.value))
		case *float64:
			*dest = yacPointerToFloat64(bind.value)
		case *string:
			*dest = C.GoString((*C.char)(bind.value))
		case *time.Time:
			timestamp := (*C.YapiTimestamp)(bind.value)
			t, tsErr := stmt.Conn.yapiTimestampToTime(timestamp, false)
			if tsErr != nil {
				return tsErr
			}
			*dest = *t
		case *bool:
			*dest = yacPointerToBool(bind.value)
		case *[]byte:
			lobLocator := (**C.YapiLobLocator)(bind.value)
			*dest, err = stmt.Conn.lobRead(*lobLocator)
			if err != nil {
				return err
			}
		case *outputBindInfo:
			switch dest.yacType {
			case C.YAPI_TYPE_BLOB:
				bindDest, _ := dest.getBlobBindDest()
				lobLocator := (**C.YapiLobLocator)(bind.value)
				*bindDest, err = stmt.Conn.lobRead(*lobLocator)
				if err != nil {
					return err
				}
			case C.YAPI_TYPE_CLOB:
				bindDest, _ := dest.getClobBindDest()
				lobLocator := (**C.YapiLobLocator)(bind.value)
				byteDest, err := stmt.Conn.lobRead(*lobLocator)
				if err != nil {
					return err
				}
				*bindDest = string(byteDest)
			case C.YAPI_TYPE_BIT:
				bindDest, _ := dest.getBitBindDest()
				if *bind.indicator != C.YAPI_NULL_DATA {
					*bindDest = C.GoBytes(unsafe.Pointer(bind.value), C.int(*bind.indicator))
				}
			case C.YAPI_TYPE_BOOL:
				bindDest, _ := dest.getBoolBindDest()
				*bindDest = yacPointerToBool(bind.value)
			case C.YAPI_TYPE_TINYINT:
				bindDest, _ := dest.getInt8BindDest()
				*bindDest = yacPointerToInt8(bind.value)
			case C.YAPI_TYPE_SMALLINT:
				bindDest, _ := dest.getInt16BindDest()
				*bindDest = yacPointerToInt16(bind.value)
			case C.YAPI_TYPE_INTEGER:
				bindDest, _ := dest.getInt32BindDest()
				*bindDest = yacPointerToInt32(bind.value)
			case C.YAPI_TYPE_BIGINT:
				bindDest, _ := dest.getInt64BindDest()
				*bindDest = yacPointerToInt64(bind.value)
			case C.YAPI_TYPE_UTINYINT:
				bindDest, _ := dest.getUint8BindDest()
				*bindDest = yacPointerToUint8(bind.value)
			case C.YAPI_TYPE_USMALLINT:
				bindDest, _ := dest.getUint16BindDest()
				*bindDest = yacPointerToUint16(bind.value)
			case C.YAPI_TYPE_UINTEGER:
				bindDest, _ := dest.getUint32BindDest()
				*bindDest = yacPointerToUint32(bind.value)
			case C.YAPI_TYPE_UBIGINT:
				bindDest, _ := dest.getUint64BindDest()
				*bindDest = yacPointerToUint64(bind.value)
			case C.YAPI_TYPE_DATE:
				bindDest, _ := dest.getTimeBindDest()
				date := yacPointerToInt64(bind.value)
				*bindDest = time.Unix(0, date*1e3).UTC()
			case C.YAPI_TYPE_TIMESTAMP, C.YAPI_TYPE_TIMESTAMP_TZ, C.YAPI_TYPE_TIMESTAMP_LTZ:
				zone := false
				if dest.yacType != C.YAPI_TYPE_TIMESTAMP {
					zone = true
				}
				bindDest, _ := dest.getTimeBindDest()
				timestamp := (*C.YapiTimestamp)(bind.value)
				t, err := stmt.Conn.yapiTimestampToTime(timestamp, zone)
				if err != nil {
					return err
				}
				*bindDest = *t
			case C.YAPI_TYPE_DOUBLE:
				bindDest, _ := dest.getFloat64BindDest()
				*bindDest = yacPointerToFloat64(bind.value)
			case C.YAPI_TYPE_FLOAT:
				bindDest, _ := dest.getFloat32BindDest()
				*bindDest = yacPointerToFloat32(bind.value)
			case C.YAPI_TYPE_YM_INTERVAL:
				bindDest, _ := dest.getIntervalBindDest()
				res, err := stmt.Conn.yapiYMIntervalToString((*C.YapiYMInterval)(bind.value))
				if err != nil {
					return err
				}
				*bindDest = res
			case C.YAPI_TYPE_DS_INTERVAL:
				bindDest, _ := dest.getIntervalBindDest()
				res, err := stmt.Conn.yapiDSIntervalToString((*C.YapiDSInterval)(bind.value))
				if err != nil {
					return err
				}
				*bindDest = res
			case C.YAPI_TYPE_BINARY:
				bindDest, _ := dest.getBlobBindDest()
				if *bind.indicator != C.YAPI_NULL_DATA {
					*bindDest = C.GoBytes(unsafe.Pointer(bind.value), C.int(*bind.indicator))
				}
			case C.YAPI_TYPE_VARCHAR, C.YAPI_TYPE_CHAR, C.YAPI_TYPE_NVARCHAR, C.YAPI_TYPE_NCHAR:
				bindDest, _ := dest.getVarcharBindDest()
				*bindDest = C.GoString((*C.char)(bind.value))
			case C.YAPI_TYPE_NUMBER:
				bindDest, _ := dest.getNumberDest()
				res, err := stmt.Conn.yapiNumberToFloat64((*C.YapiNumber)(bind.value))
				if err != nil {
					return err
				}
				*bindDest = res
			case C.YAPI_TYPE_ROWID:
				bindDest, _ := dest.getVarcharBindDest()
				*bindDest = C.GoString((*C.char)(bind.value))
			case C.YAPI_TYPE_CURSOR:
				bindDest, _ := dest.getCursorBindDest()
				handle := (*C.YacHandle)(bind.value)
				var cCursorStmt *C.YapiStmt
				if err := yapiCursorStmtCreate(stmt.Conn.Conn, &cCursorStmt, handle); err != nil {
					return err
				}
				cursorStmt := &YasStmt{
					Conn: stmt.Conn,
					Stmt: cCursorStmt,
					ctx:  stmt.ctx,
				}
				rows, err := cursorStmt.getFetchRows()
				if err != nil {
					_ = yapiReleaseCursorStmt(cCursorStmt)
					return err
				}
				bindDest.fetchRows = rows
				bindDest.isCursor = true
				bindDest.stmt = cursorStmt
			case C.YAPI_TYPE_VECTOR:
				bindDest, err := dest.getVectorBindDest()
				if err != nil {
					return err
				}
				// 从 bind.value 中读取 Vector 数据
				vec, err := stmt.Conn.vectorRead(unsafe.Pointer(bind.value))
				if err != nil {
					return err
				}
				if vec != nil {
					bindDest.Data = vec.Data
					bindDest.Dim = vec.Dim
					bindDest.Format = vec.Format
				}
			}
		default:
			return fmt.Errorf("unknown column %v", index)
		}
	}
	return nil
}

func (stmt *YasStmt) freeBindValues() {
	for _, bind := range stmt.binds {
		stmt.freeBIndValue(bind)
	}
	stmt.binds = []*bindStruct{}
}

func (stmt *YasStmt) freeBIndValue(bind *bindStruct) {
	if bind.value == nil {
		return
	}
	switch bind.freeType {
	case lobFree:
		lobLocator := (**C.YapiLobLocator)(unsafe.Pointer(bind.value))
		stmt.Conn.lobFree(bind.yacType, *lobLocator)
	case normalFree:
		C.free(unsafe.Pointer(bind.value))
	case vectorFree:
		if bind.value != nil {
			// 获取实际的 descriptor 指针
			// bind.value 指向的是一个 *unsafe.Pointer 变量，该变量包含实际的 descriptor
			ptrToDesc := (*unsafe.Pointer)(bind.value)
			actualDesc := *ptrToDesc
			if actualDesc != nil {
				// 使用包装函数，避免传递 Go 指针地址给 C
				yapiDescFree2(stmt.Conn.Env, actualDesc, C.YAPI_TYPE_VECTOR)
			}
		}
	}
	bind.value = nil
}

type outputBindInfo struct {
	yacType  C.YapiType
	dest     interface{}
	bindSize C.int32_t
}
type outputBindOpt func(*outputBindInfo)

func WithTypeBool() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_BOOL }
}

func WithTypeDate() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_DATE }
}

func WithTypeTimestamp() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_TIMESTAMP }
}

func WithTypeTimestampLocalTimeZone() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_TIMESTAMP_LTZ }
}

func WithTypeTimestampTimeZone() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_TIMESTAMP_TZ }
}

func WithTypeBigInt() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_BIGINT }
}

func WithTypeInteger() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_INTEGER }
}

func WithTypeSmallInt() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_SMALLINT }
}

func WithTypeDouble() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_DOUBLE }
}

func WithTypeFloat() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_FLOAT }
}

func WithTypeTinyint() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_TINYINT }
}

func WithTypeBit() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_BIT }
}

func WithTypeClob() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_CLOB }
}

func WithTypeBlob() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_BLOB }
}

func WithTypeVarchar() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_VARCHAR }
}

func WithTypeNvarchar() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_NVARCHAR }
}

func WithTypeChar() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_CHAR }
}

func WithTypeRaw() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_BINARY }
}

func WithTypeCursor() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_CURSOR }
}

func WithTypeDSInterval() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_DS_INTERVAL }
}

func WithTypeYMInterval() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_YM_INTERVAL }
}

func WithTypeNumber() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_NUMBER }
}

func WithTypeRowid() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_ROWID }
}

func WithCursor() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_CURSOR }
}

func WithTypeVector() outputBindOpt {
	return func(obi *outputBindInfo) { obi.yacType = C.YAPI_TYPE_VECTOR }
}

func WithBindSize(bindSize uint32) outputBindOpt {
	return func(obi *outputBindInfo) { obi.bindSize = C.int32_t(bindSize) }
}

func NewOutputBindValue(dest interface{}, opts ...outputBindOpt) (*outputBindInfo, error) {
	out := &outputBindInfo{
		dest:     dest,
		bindSize: C.int32_t(0),
		yacType:  C.YapiType(0),
	}
	if err := out.setBindOpt(opts...); err != nil {
		return nil, err
	}
	return out, nil
}

func (obi *outputBindInfo) setBindOpt(opts ...outputBindOpt) error {
	for _, opt := range opts {
		opt(obi)
	}
	return obi.checkBindOptParams()
}

func (obi *outputBindInfo) checkBindOptParams() (err error) {
	switch obi.yacType {
	case C.YAPI_TYPE_BLOB, C.YAPI_TYPE_BINARY:
		_, err = obi.getBlobBindDest()
	case C.YAPI_TYPE_CLOB:
		_, err = obi.getClobBindDest()
	case C.YAPI_TYPE_VARCHAR, C.YAPI_TYPE_CHAR, C.YAPI_TYPE_NCHAR, C.YAPI_TYPE_NVARCHAR, C.YAPI_TYPE_ROWID:
		_, err = obi.getVarcharBindDest()
	case C.YAPI_TYPE_BIT:
		_, err = obi.getBitBindDest()
	case C.YAPI_TYPE_BOOL:
		_, err = obi.getBoolBindDest()
	case C.YAPI_TYPE_DATE, C.YAPI_TYPE_TIMESTAMP, C.YAPI_TYPE_TIMESTAMP_TZ, C.YAPI_TYPE_TIMESTAMP_LTZ:
		_, err = obi.getTimeBindDest()
	case C.YAPI_TYPE_TINYINT:
		_, err = obi.getInt8BindDest()
	case C.YAPI_TYPE_SMALLINT:
		_, err = obi.getInt16BindDest()
	case C.YAPI_TYPE_BIGINT:
		_, err = obi.getInt64BindDest()
	case C.YAPI_TYPE_INTEGER:
		_, err = obi.getInt32BindDest()
	case C.YAPI_TYPE_UTINYINT:
		_, err = obi.getUint8BindDest()
	case C.YAPI_TYPE_USMALLINT:
		_, err = obi.getUint16BindDest()
	case C.YAPI_TYPE_UBIGINT:
		_, err = obi.getUint64BindDest()
	case C.YAPI_TYPE_UINTEGER:
		_, err = obi.getUint32BindDest()
	case C.YAPI_TYPE_DOUBLE:
		_, err = obi.getFloat64BindDest()
	case C.YAPI_TYPE_FLOAT:
		_, err = obi.getFloat32BindDest()
	case C.YAPI_TYPE_DS_INTERVAL, C.YAPI_TYPE_YM_INTERVAL:
		_, err = obi.getIntervalBindDest()
	case C.YAPI_TYPE_NUMBER:
		_, err = obi.getNumberDest()
	case C.YAPI_TYPE_CURSOR:
		_, err = obi.getCursorBindDest()
	case C.YAPI_TYPE_VECTOR:
		_, err = obi.getVectorBindDest()
	default:
		return ErrUnknowType(obi.yacType)
	}
	return err
}

func (obi *outputBindInfo) getClobBindDest() (*string, error) {
	if value, ok := obi.dest.(*string); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*string")
}

func (obi *outputBindInfo) getCursorBindDest() (*YasRows, error) {
	if value, ok := obi.dest.(*YasRows); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*YasRows")
}

func (obi *outputBindInfo) getBlobBindDest() (*[]byte, error) {
	if value, ok := obi.dest.(*[]byte); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*[]byte")
}

func (obi *outputBindInfo) getVarcharBindDest() (*string, error) {
	return obi.getClobBindDest()
}

func (obi *outputBindInfo) getBitBindDest() (*[]byte, error) {
	return obi.getBlobBindDest()
}

func (obi *outputBindInfo) getBoolBindDest() (*bool, error) {
	if value, ok := obi.dest.(*bool); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*bool")
}

func (obi *outputBindInfo) getInt64BindDest() (*int64, error) {
	if value, ok := obi.dest.(*int64); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*int64")
}

func (obi *outputBindInfo) getInt32BindDest() (*int32, error) {
	if value, ok := obi.dest.(*int32); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*int32")
}

func (obi *outputBindInfo) getInt16BindDest() (*int16, error) {
	if value, ok := obi.dest.(*int16); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*int16")
}

func (obi *outputBindInfo) getInt8BindDest() (*int8, error) {
	if value, ok := obi.dest.(*int8); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*int8")
}

func (obi *outputBindInfo) getUint64BindDest() (*uint64, error) {
	if value, ok := obi.dest.(*uint64); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*uint64")
}

func (obi *outputBindInfo) getUint32BindDest() (*uint32, error) {
	if value, ok := obi.dest.(*uint32); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*uint32")
}

func (obi *outputBindInfo) getUint16BindDest() (*uint16, error) {
	if value, ok := obi.dest.(*uint16); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*uint16")
}

func (obi *outputBindInfo) getUint8BindDest() (*uint8, error) {
	if value, ok := obi.dest.(*uint8); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*uint8")
}

func (obi *outputBindInfo) getTimeBindDest() (*time.Time, error) {
	if value, ok := obi.dest.(*time.Time); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*time.Time")
}

func (obi *outputBindInfo) getFloat64BindDest() (*float64, error) {
	if value, ok := obi.dest.(*float64); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*float64")
}

func (obi *outputBindInfo) getFloat32BindDest() (*float32, error) {
	if value, ok := obi.dest.(*float32); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*float32")
}

func (obi *outputBindInfo) getIntervalBindDest() (*string, error) {
	return obi.getClobBindDest()
}

func (obi *outputBindInfo) getNumberDest() (*float64, error) {
	if value, ok := obi.dest.(*float64); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*float64")
}

func (obi *outputBindInfo) getVectorBindDest() (*Vector, error) {
	if value, ok := obi.dest.(*Vector); ok {
		return value, nil
	}
	return nil, NewBindOutDestTypeErr("*yasdb.Vector")
}

func NewBindOutDestTypeErr(typeFormat string) *BindOutDestTypeErr {
	return &BindOutDestTypeErr{
		TypeFormat: typeFormat,
	}
}

type BindOutDestTypeErr struct {
	TypeFormat string
}

func (d *BindOutDestTypeErr) Error() string {
	return fmt.Sprintf("the dest parameter type must be %s", d.TypeFormat)
}

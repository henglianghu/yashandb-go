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
	"io"
	"math"
	"reflect"
	"strconv"
	"time"
	"unsafe"
)

type yasRow struct {
	Elements   uint32
	Size       uint32
	IsValueSet bool
	IsArray    bool
	Data       unsafe.Pointer
	Indicator  *C.int32_t
	yacType    C.YapiType
	name       string
	freeType   valueFreeType
	precision  uint8
	scale      int8
	nullable   uint8
	env        *C.YapiEnv // 新增：用于释放 vector 资源
}

func NewYasRow(size uint32, yacType C.YapiType, precision uint8, scale int8, nullable uint8) *yasRow {
	row := &yasRow{
		Elements:  1,
		yacType:   yacType,
		Size:      size,
		precision: precision,
		scale:     scale,
		nullable:  nullable,
	}
	return row
}

type YasRows struct {
	isCursor  bool
	stmt      *YasStmt
	fetchRows []*yasRow
	isClosed  bool
}

// Columns returns the names of the columns.
func (r *YasRows) Columns() []string {
	names := make([]string, 0, len(r.fetchRows))
	for _, row := range r.fetchRows {
		names = append(names, row.name)
	}
	return names
}

// Close closes the rows iterator.
func (r *YasRows) Close() error {
	if r.isClosed {
		return nil
	}
	freeFetchRows(r.fetchRows)
	r.isClosed = true
	if r.isCursor {
		_ = yapiReleaseCursorStmt(r.stmt.Stmt)
	}
	return nil
}

// Next is called to populate the next row of data into
// the provided slice. The provided slice will be the same
// size as the Columns() are wide.
//
// Next should return io.EOF when there are no more rows.
//
// The dest should not be written to outside of Next. Care
// should be taken when closing Rows not to modify
// a buffer held in dest.
func (r *YasRows) Next(dest []driver.Value) error {
	if r.isClosed {
		return nil
	}
	if r.stmt.ctx.Err() != nil {
		return r.stmt.ctx.Err()
	}
	r.stmt.Lock()
	defer r.stmt.Unlock()

	if r.stmt.ctx != context.Background() {
		done := make(chan struct{})
		defer close(done)
		go r.stmt.Conn.handleYacCancel(r.stmt.ctx, done)
	}

	results, err := r.getValues()
	if err != nil {
		return err
	}
	if results == nil {
		return io.EOF
	}
	copy(dest, *results)
	return nil
}

// ColumnTypeScanType return the value type that can be used to scan types into.
// For example, the database column type "bigint" this should return "reflect.TypeOf(int64(0))".
func (r *YasRows) ColumnTypeScanType(index int) reflect.Type {
	if len(r.fetchRows) < index+1 {
		return reflect.TypeOf(nil)
	}
	switch r.fetchRows[index].yacType {
	case C.YAPI_TYPE_BOOL:
		return reflect.TypeOf(false)
	case C.YAPI_TYPE_TINYINT:
		return reflect.TypeOf(int8(0))
	case C.YAPI_TYPE_SMALLINT:
		return reflect.TypeOf(int16(0))
	case C.YAPI_TYPE_INTEGER:
		return reflect.TypeOf(int32(0))
	case C.YAPI_TYPE_BIGINT:
		return reflect.TypeOf(int64(0))
	case C.YAPI_TYPE_UTINYINT:
		return reflect.TypeOf(uint8(0))
	case C.YAPI_TYPE_USMALLINT:
		return reflect.TypeOf(uint16(0))
	case C.YAPI_TYPE_UINTEGER:
		return reflect.TypeOf(uint32(0))
	case C.YAPI_TYPE_UBIGINT:
		return reflect.TypeOf(uint64(0))
	case C.YAPI_TYPE_FLOAT:
		return reflect.TypeOf(float32(0))
	case C.YAPI_TYPE_DOUBLE:
		return reflect.TypeOf(float64(0))
	case C.YAPI_TYPE_NUMBER:
		if r.stmt.Conn.numberAsString {
			return reflect.TypeOf("")
		}
		return reflect.TypeOf(float64(0))
	case C.YAPI_TYPE_DATE, C.YAPI_TYPE_TIMESTAMP, C.YAPI_TYPE_SHORTDATE, C.YAPI_TYPE_SHORTTIME, C.YAPI_TYPE_TIMESTAMP_TZ, C.YAPI_TYPE_TIMESTAMP_LTZ:
		return reflect.TypeOf(time.Time{})
	case C.YAPI_TYPE_CHAR, C.YAPI_TYPE_NCHAR, C.YAPI_TYPE_VARCHAR, C.YAPI_TYPE_NVARCHAR, C.YAPI_TYPE_CLOB, C.YAPI_TYPE_NCLOB, C.YAPI_TYPE_YM_INTERVAL, C.YAPI_TYPE_DS_INTERVAL, C.YAPI_TYPE_JSON, C.YAPI_TYPE_XML:
		return reflect.TypeOf("")
	case C.YAPI_TYPE_VECTOR:
		return reflect.TypeOf(Vector{})
	case C.YAPI_TYPE_BLOB, C.YAPI_TYPE_BINARY, C.YAPI_TYPE_BIT:
		return reflect.TypeOf([]byte(nil))
	case C.YAPI_TYPE_CURSOR:
		return reflect.TypeOf(sql.Rows{})
	default:
		return reflect.TypeOf(nil)
	}
}

// RowsColumnTypeDatabaseTypeName return the database system type name without the length. Type names should be uppercase.
func (r *YasRows) ColumnTypeDatabaseTypeName(index int) string {
	if len(r.fetchRows) < index+1 {
		return ""
	}
	return GetDatabaseTypeName(uint32(r.fetchRows[index].yacType))
}

// RowsColumnTypeLength return the length of the column type if the column is a variable length type.
// If the column is not a variable length type ok should return false.
// If length is not limited other than system limits, it should return math.MaxInt64.
func (r *YasRows) ColumnTypeLength(index int) (length int64, ok bool) {
	if len(r.fetchRows) < index+1 {
		return 0, false
	}
	switch r.fetchRows[index].yacType {
	case C.YAPI_TYPE_CHAR, C.YAPI_TYPE_NCHAR, C.YAPI_TYPE_VARCHAR, C.YAPI_TYPE_NVARCHAR:
		return int64(r.fetchRows[index].Size), true
	case C.YAPI_TYPE_BLOB, C.YAPI_TYPE_CLOB, C.YAPI_TYPE_XML:
		return math.MaxInt64, true
	default:
		return 0, false
	}
}

func (r *YasRows) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	if len(r.fetchRows) < index+1 {
		return 0, 0, false
	}
	switch r.fetchRows[index].yacType {
	case C.YAPI_TYPE_NUMBER:
		return int64(r.fetchRows[index].precision), int64(r.fetchRows[index].scale), true
	default:
		return 0, 0, false
	}
}

func (r *YasRows) ColumnTypeNullable(index int) (nullable, ok bool) {
	return r.fetchRows[index].nullable > 0, true
}

func (r *YasRows) getValues() (*[]driver.Value, error) {
	var err error
	unsafeRows := (unsafe.Pointer)(new(uint32))
	rows := (*C.uint32_t)(unsafeRows)
	if err = yapiFetch(r.stmt.Stmt, rows); err != nil {
		return nil, err
	}
	if *rows == 0 {
		return nil, nil
	}
	columns := len(r.fetchRows)
	dest := make([]driver.Value, columns)
	for i := 0; i < columns; i++ {
		row := r.fetchRows[i]
		if *row.Indicator == C.YAPI_NULL_DATA {
			dest[i] = nil
			continue
		}
		var value driver.Value
		switch row.yacType {
		case C.YAPI_TYPE_BOOL:
			value = (*(*bool)(row.Data))
		case C.YAPI_TYPE_TINYINT:
			value = (*(*int8)(row.Data))
		case C.YAPI_TYPE_SMALLINT:
			value = (*(*int16)(row.Data))
		case C.YAPI_TYPE_INTEGER:
			value = (*(*int32)(row.Data))
		case C.YAPI_TYPE_BIGINT:
			value = (*(*int64)(row.Data))
		case C.YAPI_TYPE_UTINYINT:
			value = (*(*uint8)(row.Data))
		case C.YAPI_TYPE_USMALLINT:
			value = (*(*uint16)(row.Data))
		case C.YAPI_TYPE_UINTEGER:
			value = (*(*uint32)(row.Data))
		case C.YAPI_TYPE_UBIGINT:
			value = (*(*uint64)(row.Data))
		case C.YAPI_TYPE_FLOAT:
			value = (*(*float32)(row.Data))
		case C.YAPI_TYPE_DOUBLE:
			value = (*(*float64)(row.Data))
		case C.YAPI_TYPE_DATE:
			tmpDate := time.UnixMicro(*(*int64)(row.Data)).UTC()
			value = time.Date(tmpDate.Year(), tmpDate.Month(), tmpDate.Day(), tmpDate.Hour(), tmpDate.Minute(), tmpDate.Second(), 0, time.UTC)
		case C.YAPI_TYPE_TIMESTAMP:
			ts := (*C.YapiTimestamp)(row.Data)
			t, err := r.stmt.Conn.yapiTimestampToTime(ts, false)
			if err != nil {
				return nil, err
			}
			value = *t
		case C.YAPI_TYPE_SHORTDATE:
			tmpDate := time.UnixMicro(*(*int64)(row.Data)).UTC()
			value = time.Date(tmpDate.Year(), tmpDate.Month(), tmpDate.Day(), 0, 0, 0, 0, time.UTC)
		case C.YAPI_TYPE_SHORTTIME:
			tmpDate := time.UnixMicro(*(*int64)(row.Data)).UTC()
			value = time.Date(0, 1, 1, tmpDate.Hour(), tmpDate.Minute(), tmpDate.Second(), tmpDate.Nanosecond(), time.UTC)
		case C.YAPI_TYPE_TIMESTAMP_LTZ:
			ts := (*C.YapiTimestamp)(row.Data)
			t, err := r.stmt.Conn.yapiTimestampToTime(ts, true)
			if err != nil {
				return nil, err
			}
			value = *t
		case C.YAPI_TYPE_TIMESTAMP_TZ:
			valueStr := fixTimeString(C.GoString((*C.char)(row.Data)))
			t, err := time.Parse(DateTimeFormats[DateTimeMicroZone], valueStr)
			if err != nil {
				return nil, fmt.Errorf("convert %q to time.Time failed, %v", valueStr, err)
			}
			value = t
		case C.YAPI_TYPE_CHAR, C.YAPI_TYPE_NCHAR, C.YAPI_TYPE_VARCHAR, C.YAPI_TYPE_NVARCHAR:
			value = (C.GoString((*C.char)(row.Data)))
		case C.YAPI_TYPE_NUMBER, C.YAPI_TYPE_NUMBER_FLOAT:
			str := C.GoString((*C.char)(row.Data))
			if r.stmt.Conn.numberAsString {
				value = str
			} else {
				value, err = strconv.ParseFloat(str, 64)
				if err != nil {
					return nil, err
				}
			}
		case C.YAPI_TYPE_CLOB, C.YAPI_TYPE_BLOB, C.YAPI_TYPE_XML, C.YAPI_TYPE_NCLOB:
			lobLocator := (**C.YapiLobLocator)(row.Data)
			data, err := r.stmt.Conn.lobRead(*lobLocator)
			if err != nil {
				return nil, err
			}
			switch row.yacType {
			case C.YAPI_TYPE_CLOB, C.YAPI_TYPE_XML, C.YAPI_TYPE_NCLOB:
				value = string(data)
			default:
				value = data
			}
		case C.YAPI_TYPE_BINARY:
			data := (*[65535]byte)(row.Data)[0:*row.Indicator]
			value = data
		case C.YAPI_TYPE_BIT:
			data := (*[64]byte)(row.Data)[0:*row.Indicator]
			value = data
		case C.YAPI_TYPE_YM_INTERVAL:
			cymi := (*C.YapiYMInterval)(row.Data)
			ymi, err := r.stmt.Conn.yapiYMIntervalToYmInterval(*cymi)
			if err != nil {
				return nil, err
			}
			value = ymi
		case C.YAPI_TYPE_DS_INTERVAL:
			cdsi := (*C.YapiDSInterval)(row.Data)
			dsi, err := r.stmt.Conn.yapiDSIntervalToDSInterval(*cdsi)
			if err != nil {
				return nil, err
			}
			value = dsi
		case C.YAPI_TYPE_CURSOR:
			handle := (*C.YacHandle)(row.Data)
			var cCursorStmt *C.YapiStmt
			if err := yapiCursorStmtCreate(r.stmt.Conn.Conn, &cCursorStmt, handle); err != nil {
				return nil, err
			}
			cursorStmt := &YasStmt{
				Conn: r.stmt.Conn,
				Stmt: cCursorStmt,
				ctx:  r.stmt.ctx,
			}
			rows, err := cursorStmt.getFetchRows()
			if err != nil {
				_ = yapiReleaseCursorStmt(cCursorStmt)
				return nil, err
			}
			value = &YasRows{
				stmt:      cursorStmt,
				fetchRows: rows,
				isCursor:  true,
			}

		case C.YAPI_TYPE_VECTOR:
			vector, err := r.stmt.Conn.vectorRead(row.Data)
			if err != nil {
				return nil, err
			}
			value = vector
		default:
			value = (C.GoString((*C.char)(row.Data)))
		}
		dest[i] = value
	}
	return &dest, nil
}

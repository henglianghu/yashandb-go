/*
Copyright  2022, YashanDB and/or its affiliates. All rights reserved.
YashanDB Driver for golang is licensed under the terms of the mulan PSL v2.0

License: 	http://license.coscl.org.cn/MulanPSL2
Home page: 	https://www.yashandb.com/
*/

package yasdb

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const (
	_UnknownAttributeIdCode = "YAS-08028"
	_LoadSymbolCode         = "YAS-20001"
)

type YasBaseError struct {
	Code int
	Msg  string
}

func (e *YasBaseError) Error() string { return e.Msg }

type YasDBError struct {
	Code     int
	Msg      string
	SqlState string
	Line     int
	Column   int
}

func (e *YasDBError) Error() string {
	if e.Code == 0 {
		if len(e.Msg) != 0 {
			return e.Msg
		}
		return "yasdb return code is zero"
	}

	if e.Code == -1 {
		if e.Line > 0 || e.Column > 0 {
			return fmt.Sprintf("[%d:%d]YAS%d %s", e.Line, e.Column, e.Code, e.Msg)
		}
		return fmt.Sprintf("YAS%d %s", e.Code, e.Msg)
	}

	if e.Line > 0 || e.Column > 0 {
		return fmt.Sprintf("[%d:%d]YAS-%05d %s", e.Line, e.Column, e.Code, e.Msg)
	}
	return fmt.Sprintf("YAS-%05d %s", e.Code, e.Msg)
}

func ErrDsnNoStandard(dsn string) *YasBaseError {
	return &YasBaseError{Code: 1001, Msg: "dsn is nonstandard"}
}

func ErrDsnNoSet() *YasBaseError {
	return &YasBaseError{Code: 1002, Msg: "dsn is unset"}
}

func ErrDataPathNoExist(p string) *YasBaseError {
	return &YasBaseError{Code: 1009, Msg: fmt.Sprintf("YASDB_DATA:%s is not existed", p)}
}

func ErrEnvInit() *YasBaseError {
	return &YasBaseError{Code: 1005, Msg: "yasdb env init failed"}
}

func ErrNoConnect() *YasBaseError {
	return &YasBaseError{Code: 1003, Msg: "yasdb is not connected"}
}

func ErrStmtInit() *YasBaseError {
	return &YasBaseError{Code: 1004, Msg: "yasdb stmt init failed"}
}

func ErrStmtNoOpen() *YasBaseError {
	return &YasBaseError{Code: 1005, Msg: "yasdb stmt is not open"}
}

func ErrDbTypeUnsupport(dbType int) *YasBaseError {
	return &YasBaseError{Code: 1006, Msg: fmt.Sprintf("yasdb type %d is not support", dbType)}
}

func ErrDbFetchEOF() *YasBaseError {
	return &YasBaseError{Code: 1007, Msg: "fetch is over"}
}

func ErrUnknowType(v interface{}) *YasBaseError {
	return &YasBaseError{Code: 1008, Msg: fmt.Sprintf("Unknow: %s", reflect.TypeOf(v).String())}
}

func ErrNoCursorHandle() *YasBaseError {
	return &YasBaseError{Code: 1009, Msg: "yasdb cursor handle is null"}
}

func ErrOutputBindValue() *YasBaseError {
	return &YasBaseError{}
}

func ErrVectorNil() *YasBaseError {
	return &YasBaseError{Code: 1010, Msg: "vector is nil"}
}

func isUnknownAttributeIdErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), _UnknownAttributeIdCode)
}

func isLoadSymbolErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), _LoadSymbolCode)
}

func isFeatureUnSupported(err error) bool {
	if err == nil {
		return false
	}
	var yasErr *YasDBError
	if !errors.As(err, &yasErr) {
		return false
	}
	return yasErr.Code == 8052
}

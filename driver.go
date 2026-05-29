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
	"database/sql/driver"
	"unsafe"
)

var _ClientDriverName = "YashanDB GO Driver"

func SetClientDriverName(name string) {
	if name == "" {
		return
	}
	_ClientDriverName = name
}

const (
	_DebugMode   = "debug"
	_ReleaseMode = "release"
)

const (
	_DefaultDbTimestampFormat   = "yyyy-mm-dd hh24:mi:ss.ff"
	_DefaultDbDateFormat        = "yyyy-mm-dd"
	_DefaultDbTimeFormat        = "hh24:mi:ss.ff"
	_DefaultDbTimestampTzFormat = "yyyy-mm-dd hh24:mi:ss.ff tzh:tzm"
	_DefaultDbDsIntervalFormat  = "dd hh24:mi:ss.ff"
	_DefaultDbYmIntervalFormat  = "yy-mm"
)

var _Mode = _ReleaseMode

func SetDebugMode() {
	_Mode = _DebugMode
}

type YasdbDriver struct {
	conn *YasConn
}

// Open returns a new connection to the database.
func (yasDriver *YasdbDriver) Open(dsnStr string) (driver.Conn, error) {
	conn, err := GenYasconn(dsnStr)
	if err != nil {
		return nil, err
	}
	yasDriver.conn = conn
	return conn, nil
}

func (yasDriver *YasdbDriver) Conn() *YasConn {
	// 需要实际连接才会有conn，仅Open时为nil
	return yasDriver.conn
}

func GenYasconn(dsnStr string) (*YasConn, error) {
	dsn, err := ParseDSN(dsnStr)
	if err != nil {
		return nil, err
	}

	var env *C.YapiEnv
	if err := yapiAllocEnv(&env); err != nil {
		return nil, err
	}
	if dsn.DataPath != "" {
		dataPath := stringToYasChar(dsn.DataPath)
		defer C.free(unsafe.Pointer(dataPath))

		dpLen := intToYacInt32(len(dsn.DataPath))
		if err := yapiSetEnvAttr(env, C.YAPI_ATTR_DATA_PATH, unsafe.Pointer(dataPath), dpLen); err != nil {
			_ = releaseEnv(env)
			return nil, err
		}
	}

	charset := C.YAPI_CHARSET_UTF8
	if err := yapiSetEnvAttr(env, C.YAPI_ATTR_CHARSET_CODE, unsafe.Pointer(&charset), 4); err != nil {
		_ = releaseEnv(env)
		return nil, err
	}

	driverName := stringToYasChar(_ClientDriverName)
	driverNameLen := intToYacInt32(len(_ClientDriverName))
	defer C.free(unsafe.Pointer(driverName))
	C.yapiSetEnvAttr(env, C.YAPI_ATTR_CLIENT_DRIVER, unsafe.Pointer(driverName), driverNameLen)

	var conn *C.YapiConnect

	url := stringToYasChar(dsn.Url)
	defer C.free(unsafe.Pointer(url))
	user := stringToYasChar(dsn.User)
	defer C.free(unsafe.Pointer(user))
	password := stringToYasChar(dsn.Password)
	defer C.free(unsafe.Pointer(password))
	urlLen := intToYacInt16(len(dsn.Url))
	userLen := intToYacInt16(len(dsn.User))
	pwLen := intToYacInt16(len(dsn.Password))

	if err := yapiAllocConnect(env, &conn); err != nil {
		_ = releaseEnv(env)
		return nil, err
	}

	yasConn := &YasConn{
		Env:               env,
		Conn:              conn,
		numberAsString:    dsn.numberAsString,
		cliPrepare:        dsn.cliPrepare,
		autocommit:        dsn.IsAutoCommit,
		dateFormat:        dsn.dateFormat,
		timeFormat:        dsn.timeFormat,
		timestampFormat:   dsn.timestampFormat,
		timestampTzFormat: dsn.timestampTzFormat,
		dsIntervalFormat:  dsn.dsIntervalFormat,
		ymIntervalFormat:  dsn.ymIntervalFormat,
		compatVector:      dsn.compatVector,
	}

	if err := yasConn.setCompatVector(dsn.compatVector); err != nil {
		_ = yasConn.Close()
		return nil, err
	}

	if err := yasConn.setSslRootCer(dsn.sslRootCer); err != nil {
		_ = yasConn.Close()
		return nil, err
	}

	if err := yasConn.setHeartbeatEnable(dsn.heartbeatEnable); err != nil {
		_ = yasConn.Close()
		return nil, err
	}

	if err := yapiConnect2(conn, url, urlLen, user, userLen, password, pwLen); err != nil {
		_ = releaseEnv(env)
		return nil, err
	}

	if err := yasConn.setAutoCommit(dsn.IsAutoCommit); err != nil {
		_ = yasConn.Close()
		return nil, err
	}
	yasConn.autocommit = dsn.IsAutoCommit

	if err := yasConn.getConnAttr(); err != nil {
		_ = yasConn.Close()
		return nil, err
	}

	return yasConn, nil
}

type YasTx struct {
	Conn *YasConn
}

// Commit transaction commit
func (tx *YasTx) Commit() error {
	return tx.Conn.yacCommit()
}

// Rollback transaction rollback
func (tx *YasTx) Rollback() error {
	return tx.Conn.yacRollback()
}

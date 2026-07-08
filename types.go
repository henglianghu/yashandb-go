package yasdb

/*
#cgo CFLAGS: -I./yacapi/include -I./yacapi/src

#include "yacapi.h"
#include "yapi_inc.h"
#include <stdio.h>
#include <stdlib.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

func yapiReleaseConn(conn *C.YapiConnect) error {
	if conn == nil {
		return nil
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiReleaseConn(conn))
}

func yapiReleaseEnv(env *C.YapiEnv) error {
	if env == nil {
		return nil
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiReleaseEnv(env))
}

func yapiAllocEnv(env **C.YapiEnv) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiAllocEnv(env))
}

func yapiSetConnAttr(conn *C.YapiConnect, attr C.YapiConnAttr, value unsafe.Pointer, bufLength C.int32_t) error {
	if conn == nil {
		return ErrNoConnect()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiSetConnAttr(
		conn,
		attr,
		value,
		bufLength,
	))
}

func yapiGetConnAttr(conn *C.YapiConnect, attr C.YapiConnAttr, value unsafe.Pointer, bufLength C.int32_t) error {
	if conn == nil {
		return ErrNoConnect()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var stringLen C.int32_t
	return checkYasError(
		C.yapiGetConnAttr(
			conn,
			attr,
			value,
			bufLength,
			&stringLen,
		))
}

func yapiCommit(conn *C.YapiConnect) error {
	if conn == nil {
		return ErrNoConnect()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiCommit(conn))
}

func yapiRollback(conn *C.YapiConnect) error {
	if conn == nil {
		return ErrNoConnect()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiRollback(conn))
}

func yapiLobRead(conn *C.YapiConnect, lobLocator *C.YapiLobLocator, bytes *C.uint64_t, buf *C.uint8_t) error {
	if conn == nil {
		return ErrNoConnect()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(
		C.yapiLobRead(
			conn,
			lobLocator,
			bytes,
			buf,
			_LobBufLen,
		))
}

func yapiLobGetLength(conn *C.YapiConnect, lobLocator *C.YapiLobLocator, lobLen *C.uint64_t) error {
	if conn == nil {
		return ErrNoConnect()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiLobGetLength(conn, lobLocator, lobLen))
}

func yapiLobDescAlloc(conn *C.YapiConnect, yacType C.YapiType, desc *unsafe.Pointer) error {
	if conn == nil {
		return ErrNoConnect()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiLobDescAlloc(conn, yacType, desc))
}

func yapiLobCreateTemporary(conn *C.YapiConnect, lobLocator *C.YapiLobLocator) error {
	if conn == nil {
		return ErrNoConnect()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiLobCreateTemporary(conn, lobLocator))
}

func yapiLobWrite(conn *C.YapiConnect, lobLocator *C.YapiLobLocator, buf *C.uint8_t, bufLen C.uint64_t) error {
	if conn == nil {
		return ErrNoConnect()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(
		C.yapiLobWrite(
			conn,
			lobLocator,
			nil,
			buf,
			bufLen,
		))
}

func yapiCancel(conn *C.YapiConnect) error {
	if conn == nil {
		return ErrNoConnect()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiCancel(conn))
}

func yapiPrepare(conn *C.YapiConnect, queryP *C.char, sqlLength C.int32_t, stmt **C.YapiStmt) error {
	if conn == nil {
		return ErrNoConnect()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(
		C.yapiPrepare(
			conn,
			queryP,
			sqlLength,
			stmt,
		))
}

func yapiGetStmtAttr(stmt *C.YapiStmt, stmtAttr C.YapiStmtAttr, point unsafe.Pointer, sqlSize, sqlLength C.int32_t) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(
		C.yapiGetStmtAttr(
			stmt,
			stmtAttr,
			point,
			sqlSize,
			&sqlLength,
		))
}

func yapiExecute(stmt *C.YapiStmt) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiExecute(stmt))
}

func yapiDirectExecute(stmt *C.YapiStmt, sqlstr string) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	queryP := C.CString(sqlstr)
	defer C.free(unsafe.Pointer(queryP))
	sqlLength := C.int32_t(len(sqlstr))
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiDirectExecute(stmt, queryP, sqlLength))
}

func yapiStmtCreate(conn *C.YapiConnect, stmt **C.YapiStmt) error {
	if conn == nil {
		return ErrNoConnect()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiStmtCreate(conn, stmt))
}

func yapiCursorStmtCreate(conn *C.YapiConnect, stmt **C.YapiStmt, handle *C.YacHandle) error {
	if conn == nil {
		return ErrNoConnect()
	}
	if handle == nil {
		return ErrNoCursorHandle()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiCursorStmtCreate(conn, stmt, handle))
}

func yapiReleaseCursorStmt(stmt *C.YapiStmt) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiReleaseCursorStmt(stmt))
}

func yapiReleaseStmt(stmt *C.YapiStmt) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiReleaseStmt(stmt))
}

func yapiReleaseHandle(handle *C.YacHandle) error {
	if handle == nil {
		return ErrNoCursorHandle()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiReleaseHandle(handle))
}

func yapiNumResultCols(stmt *C.YapiStmt, columns *C.int16_t) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiNumResultCols(stmt, columns))
}

func yapiDescribeCol2(stmt *C.YapiStmt, pos C.uint16_t, item *C.YapiColumnDesc) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiDescribeCol2(stmt, pos, item))
}

func yapiBindColumn(stmt *C.YapiStmt, pos C.uint16_t, yacType C.YapiType, point C.YapiPointer, bufLen C.int32_t, indicator *C.int32_t) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(
		C.yapiBindColumn(
			stmt,
			pos,
			yacType,
			point,
			bufLen,
			indicator,
		),
	)
}

func yapiBindParameter(stmt *C.YapiStmt, b *bindStruct, pos C.uint16_t) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(
		C.yapiBindParameter(
			stmt,
			pos,
			b.direction,
			b.yacType,
			b.value,
			b.bindSize,
			C.int32_t(0),
			b.indicator,
		),
	)
}

func yapiBindParameterByName(stmt *C.YapiStmt, charName *C.char, b *bindStruct) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(
		C.yapiBindParameterByName(
			stmt,
			charName,
			b.direction,
			b.yacType,
			b.value,
			b.bindSize,
			C.int32_t(0),
			b.indicator,
		),
	)
}

func yapiSetEnvAttr(env *C.YapiEnv, envAttr C.YapiEnvAttr, point unsafe.Pointer, dpLen C.int32_t) error {
	if env == nil {
		return ErrEnvInit()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiSetEnvAttr(env, envAttr, point, dpLen))
}

func yapiConnect(env *C.YapiEnv, url *C.char, urlLen C.int16_t, user *C.char, userLen C.int16_t, password *C.char, pwLen C.int16_t, conn **C.YapiConnect) error {
	if env == nil {
		return ErrEnvInit()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiConnect(env, url, urlLen, user, userLen, password, pwLen, conn))
}

func yapiAllocConnect(env *C.YapiEnv, conn **C.YapiConnect) error {
	if env == nil {
		return ErrEnvInit()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiAllocConnect(env, conn))
}

func yapiConnect2(conn *C.YapiConnect, url *C.char, urlLen C.int16_t, user *C.char, userLen C.int16_t, password *C.char, pwLen C.int16_t) error {
	if conn == nil {
		return ErrNoConnect()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiConnect2(conn, url, urlLen, user, userLen, password, pwLen))
}

func yapiPdbgStart(stmt *C.YapiStmt, objId C.uint64_t, subId C.uint16_t) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgStart(stmt, objId, subId))
}

func yapiPdbgCheckVersion(stmt *C.YapiStmt, objId C.uint64_t, subId C.uint16_t, version C.uint32_t) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgCheckVersion(stmt, objId, subId, version))
}

func yapiPdbgAbort(stmt *C.YapiStmt) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgAbort(stmt))
}

func yapiPdbgContinue(stmt *C.YapiStmt) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgContinue(stmt))
}

func yapiPdbgStepInto(stmt *C.YapiStmt) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgStepInto(stmt))
}

func yapiPdbgStepOut(stmt *C.YapiStmt) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgStepOut(stmt))
}

func yapiPdbgStepNext(stmt *C.YapiStmt) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgStepNext(stmt))
}

func yapiPdbgDeleteAllBreakpoints(stmt *C.YapiStmt) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgDeleteAllBreakpoints(stmt))
}

func yapiPdbgAddBreakpoint(stmt *C.YapiStmt, objId uint64, subId uint16, lineNo uint32, bpID *C.uint32_t) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgAddBreakpoint(stmt, C.uint64_t(objId), C.uint16_t(subId), C.uint32_t(lineNo), bpID))
}

func yapiPdbgDeleteBreakpoint(stmt *C.YapiStmt, objId uint64, subId uint16, lineNo uint32) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgDeleteBreakpoint(stmt, C.uint64_t(objId), C.uint16_t(subId), C.uint32_t(lineNo)))
}

func yapiPdbgGetBreakpointsCount(stmt *C.YapiStmt, bpCount *C.uint32_t) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgGetBreakpointsCount(stmt, bpCount))
}

func yapiPdbgGetAllVars(stmt *C.YapiStmt, varCount *C.uint32_t) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgGetAllVars(stmt, varCount))
}

func yapiPdbgGetAllFrames(stmt *C.YapiStmt, frameCount *C.uint32_t) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgGetAllFrames(stmt, frameCount))
}

func yapiPdbgGetRunningAttrs(stmt *C.YapiStmt, attr C.YapiDebugRunningAttr, point C.YapiPointer, len C.int32_t, stringLen *C.int32_t) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgGetRunningAttrs(stmt, attr, point, len, stringLen))
}

func yapiPdbgGetFrameAttrs(stmt *C.YapiStmt, id C.uint32_t, attr C.YapiDebugFrameAttr, point C.YapiPointer, len C.int32_t, stringLen *C.int32_t) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgGetFrameAttrs(stmt, id, attr, point, len, stringLen))
}

func yapiPdbgGetVarAttrs(stmt *C.YapiStmt, id C.uint32_t, attr C.YapiDebugVarAttr, point C.YapiPointer, len C.int32_t, stringLen *C.int32_t) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgGetVarAttrs(stmt, id, attr, point, len, stringLen))
}

func yapiPdbgGetVarValue(stmt *C.YapiStmt, id C.uint32_t, bindType C.uint32_t, point C.YapiPointer, len C.int32_t, stringLen *C.int32_t) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgGetVarValue(stmt, id, bindType, point, len, stringLen))
}

func yapiPdbgGetBreakpointAttrs(stmt *C.YapiStmt, id C.uint32_t, attr C.YapiDebugBpAttr, point C.YapiPointer, len C.int32_t, stringLen *C.int32_t) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgGetBreakpointAttrs(stmt, C.uint32_t(id), attr, point, len, stringLen))
}

func yapiPdbgGetOutput(stmt *C.YapiStmt, buffer *C.char, length *C.uint32_t, hasMore *C.bool) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPdbgGetOutput(stmt, buffer, length, hasMore))
}

func yapiFetch(stmt *C.YapiStmt, rows *C.uint32_t) error {
	if stmt == nil {
		return ErrStmtNoOpen()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiFetch(stmt, rows))
}

func yapiTimestampGetTimestamp(timestamp C.YapiTimestamp,
	year *C.int16_t,
	month *C.uint8_t,
	day *C.uint8_t,
	hour *C.uint8_t,
	minute *C.uint8_t,
	second *C.uint8_t,
	fraction *C.uint32_t) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiTimestampGetTimestamp(timestamp, year, month, day, hour, minute, second, fraction))
}

func yapiTimestampSetTimestamp(timestamp *C.YapiTimestamp,
	year C.int16_t,
	month C.uint8_t,
	day C.uint8_t,
	hour C.uint8_t,
	minute C.uint8_t,
	second C.uint8_t,
	fraction C.uint32_t) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiTimestampSetTimestamp(timestamp, year, month, day, hour, minute, second, fraction))
}

func yapiDateTimeGetTimeZoneOffset(env *C.YapiEnv, timestamp C.YapiTimestamp, hr *C.int8_t, mm *C.int8_t) error {
	if env == nil {
		return ErrEnvInit()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiDateTimeGetTimeZoneOffset(env, timestamp, hr, mm))
}

func yapiDSIntervalFromText(hEnv *C.YapiEnv, dsInterval *C.YapiDSInterval, str *C.char, strLen C.uint32_t) error {
	if hEnv == nil {
		return ErrEnvInit()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiDSIntervalFromText(hEnv, dsInterval, str, strLen))
}

func yapiYMIntervalFromText(hEnv *C.YapiEnv, dsInterval *C.YapiYMInterval, str *C.char, strLen C.uint32_t) error {
	if hEnv == nil {
		return ErrEnvInit()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiYMIntervalFromText(hEnv, dsInterval, str, strLen))
}

func yapiDSIntervalGetDaySecond(dsInterval C.YapiDSInterval, day *C.int32_t, hour *C.int32_t, mintue *C.int32_t, second *C.int32_t, fraction *C.int32_t) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiDSIntervalGetDaySecond(dsInterval, day, hour, mintue, second, fraction))
}

func yapiDSIntervalSetDaySecond(dsInterval *C.YapiDSInterval, day C.int32_t, hour C.int32_t, mintue C.int32_t, second C.int32_t, fraction C.int32_t) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiDSIntervalSetDaySecond(dsInterval, day, hour, mintue, second, fraction))
}

func yapiYMIntervalGetYearMonth(ymInterval C.YapiYMInterval, year *C.int32_t, month *C.int32_t) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiYMIntervalGetYearMonth(ymInterval, year, month))
}

func yapiYMIntervalSetYearMonth(ymInterval *C.YapiYMInterval, year C.int32_t, month C.int32_t) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiYMIntervalSetYearMonth(ymInterval, year, month))
}

func yapiNumberToText(
	number *C.YapiNumber,
	fmt *C.char,
	fmtLength C.uint32_t,
	nlsParam *C.char,
	nlsParamLength C.uint32_t,
	str *C.char,
	bufLength C.int32_t,
	length *C.int32_t,
) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiNumberToText(number, fmt, fmtLength, nlsParam, nlsParamLength, str, bufLength, length))
}

func yapiNumberFromText(
	str *C.char,
	strLength C.uint32_t,
	fmt *C.char,
	fmtLength C.uint32_t,
	nlsParam *C.char,
	nlsParamLength C.uint32_t,
	number *C.YapiNumber,
) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiNumberFromText(str, strLength, fmt, fmtLength, nlsParam, nlsParamLength, number))
}

func yapiNumberFromReal(rnum C.YapiPointer, length C.uint32_t, number *C.YapiNumber) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiNumberFromReal(rnum, length, number))
}

func yapiNumberToReal(number *C.YapiNumber, length C.uint32_t, rsl C.YapiPointer) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiNumberToReal(number, length, rsl))
}

func yapiPing(conn *C.YapiConnect, timeout C.int32_t) error {
	// timeout is ms
	if conn == nil {
		return ErrNoConnect()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiPing(conn, timeout))
}

func yapiParseSqlParams(env *C.YapiEnv, paramList *C.YapiPointer, sql *C.char, sqlLength C.int32_t) error {
	if env == nil {
		return ErrEnvInit()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiParseSqlParams(env, paramList, sql, sqlLength))
}

func yapiGetParamListCount(hParamList C.YapiPointer, count *C.uint32_t) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiGetParamListCount(hParamList, count))
}

// func yapiGetParamName(hParamList C.YapiPointer, index C.uint16_t, name *C.char, nameBufLen C.int32_t, nameLen *C.int32_t) error {
// 	runtime.LockOSThread()
// 	defer runtime.UnlockOSThread()
// 	return checkYasError(C.yapiGetParamName(hParamList, index, name, nameBufLen, nameLen))
// }

func yapiFreeParamList(hParamList C.YapiPointer) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiFreeParamList(hParamList))
}

func yapiDescAlloc2(env *C.YapiEnv, desc *unsafe.Pointer, descType C.YapiDescType) error {
	if env == nil {
		return ErrEnvInit()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiDescAlloc2(env, desc, descType))
}

func yapiDescFree2(env *C.YapiEnv, desc unsafe.Pointer, descType C.YapiDescType) error {
	if env == nil {
		return ErrEnvInit()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiDescFree2(env, desc, descType))
}

func yapiVectorToArray(vector *C.YapiVector, format C.YapiVectorFormat, dim *C.uint16_t, array *C.uint8_t, arrayLen *C.uint32_t, mode C.uint32_t) error {
	if vector == nil {
		return ErrVectorNil()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiVectorToArray(vector, format, dim, array, arrayLen, mode))
}

func yapiVectorToText(vector *C.YapiVector, text *C.char, textlen *C.uint32_t, mode C.uint32_t) error {
	if vector == nil {
		return ErrVectorNil()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiVectorToText(vector, text, textlen, mode))
}

func yapiVectorFromText(vector *C.YapiVector, format C.YapiVectorFormat, dim C.uint16_t, text *C.char, textlen C.uint32_t, mode C.uint32_t) error {
	if vector == nil {
		return ErrVectorNil()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiVectorFromText(vector, format, dim, text, textlen, mode))
}

func yapiVectorFromArray(vector *C.YapiVector, format C.YapiVectorFormat, dim C.uint16_t, array *C.uint8_t, arrayLen C.uint32_t, mode C.uint32_t) error {
	if vector == nil {
		return ErrVectorNil()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiVectorFromArray(vector, format, dim, array, arrayLen, mode))
}

func yapiVectorGetFormat(vector *C.YapiVector, format *C.YapiVectorFormat) error {
	if vector == nil {
		return ErrVectorNil()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiVectorGetFormat(vector, format))
}

func yapiVectorGetDimension(vector *C.YapiVector, dim *C.uint16_t) error {
	if vector == nil {
		return ErrVectorNil()
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return checkYasError(C.yapiVectorGetDimension(vector, dim))
}

// yapiColumnDescGetVectorFormat returns the vector format from column descriptor.
// The C function does not fail, so no error is returned.
func yapiColumnDescGetVectorFormat(desc *C.YapiColumnDesc) uint8 {
	if desc == nil {
		return 0
	}
	return uint8(C.yapiColumnDescGetVectorFormat(desc))
}

// yapiColumnDescGetPrecision returns the precision from column descriptor.
func yapiColumnDescGetPrecision(desc *C.YapiColumnDesc) uint8 {
	if desc == nil {
		return 0
	}
	return uint8(C.yapiColumnDescGetPrecision(desc))
}

// yapiColumnDescGetScale returns the scale from column descriptor.
func yapiColumnDescGetScale(desc *C.YapiColumnDesc) int8 {
	if desc == nil {
		return 0
	}
	return int8(C.yapiColumnDescGetScale(desc))
}

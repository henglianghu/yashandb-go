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
	"errors"
	"fmt"
	"strings"
	"unsafe"
)

type DebugRunningAttr int8
type DebugFrameAttr int8
type DebugVarAttr int8
type DebugBpAttr int8
type DebuggerStatus int8

const PdbgOutputDefaultBufSize = 65534*2 + 1

const (
	DBG_RUNNING_ATTR_STATUS      DebugRunningAttr = 0
	DBG_RUNNING_ATTR_OBJ_ID      DebugRunningAttr = 1
	DBG_RUNNING_ATTR_SUB_ID      DebugRunningAttr = 2
	DBG_RUNNING_ATTR_LINE_NO     DebugRunningAttr = 3
	DBG_RUNNING_ATTR_CLASS_NAME  DebugRunningAttr = 4
	DBG_RUNNING_ATTR_METHOD_NAME DebugRunningAttr = 5

	DBG_FRAME_ATTR_OBJ_ID      DebugFrameAttr = 0
	DBG_FRAME_ATTR_SUB_ID      DebugFrameAttr = 1
	DBG_FRAME_ATTR_LINE_NO     DebugFrameAttr = 2
	DBG_FRAME_ATTR_STACK_NO    DebugFrameAttr = 3
	DBG_FRAME_ATTR_CLASS_NAME  DebugFrameAttr = 4
	DBG_FRAME_ATTR_METHOD_NAME DebugFrameAttr = 5

	DBG_VAR_ATTR_BLOCK_NO   DebugVarAttr = 0
	DBG_VAR_ATTR_TYPE       DebugVarAttr = 1
	DBG_VAR_ATTR_IS_GLOBAL  DebugVarAttr = 2
	DBG_VAR_ATTR_NAME       DebugVarAttr = 3
	DBG_VAR_ATTR_VALUE_SIZE DebugVarAttr = 4

	DBG_BP_ATTR_OBJ_ID  DebugBpAttr = 0
	DBG_BP_ATTR_SUB_ID  DebugBpAttr = 1
	DBG_BP_ATTR_LINE_NO DebugBpAttr = 2

	DBG_STATUS_OFF DebuggerStatus = 0
	DBG_STATUS_ON  DebuggerStatus = 1

	DEFAULT_ATTRS_BUFFER = 68*2 + 1
)

var (
	ValueTypeUint64Err = errors.New("the value parameter type must be *uint64")
	ValueTypeUint16Err = errors.New("the value parameter type must be *uint16")
	ValueTypeUint32Err = errors.New("the value parameter type must be *uint32")
	ValueTypeStringErr = errors.New("the value parameter type must be *string")
)

type PlsqlDebug struct {
	conn          *YasConn
	Stmt          *YasStmt
	enableStmt    *YasStmt
	getLineStmt   *YasStmt
	ctx           context.Context
	config        *debugConfig
	line          string
	getLineStatus int32
	outputBuf     *C.char
}

type debugConfig struct {
	enableOutput     bool
	outputCacheSize  uint32
	outputMaxLineLen uint32
	callTemplate     string
	callArgs         []any
}

type DebugConfigOpt func(*debugConfig)

func WithDebugCallTempalate(call string, args ...any) DebugConfigOpt {
	return func(dc *debugConfig) {
		dc.callTemplate = call
		dc.callArgs = append(dc.callArgs, args...)
	}
}

func WithDebugEnableOutput() DebugConfigOpt {
	return func(dc *debugConfig) {
		dc.enableOutput = true
	}
}

func WithDebugOutputCacheSize(cacheSize uint32) DebugConfigOpt {
	return func(dc *debugConfig) {
		dc.outputCacheSize = cacheSize
	}
}

func WithDebugOutputMaxLineLen(maxLineLen uint32) DebugConfigOpt {
	return func(dc *debugConfig) {
		dc.outputMaxLineLen = maxLineLen
	}
}

func defaultDebugConfig() *debugConfig {
	return &debugConfig{
		enableOutput:     false,
		outputCacheSize:  1000000,
		outputMaxLineLen: 65534,
	}
}

// newPlsqlDebug 是核心初始化逻辑（私有），conn 由调用方提供。
// 不在错误路径上关闭 conn —— 共享连接场景下调用方负责管理 conn 生命周期。
func newPlsqlDebug(conn *YasConn, opts ...DebugConfigOpt) (*PlsqlDebug, error) {
	config := defaultDebugConfig()
	for _, opt := range opts {
		opt(config)
	}
	if err := validateDebugConfig(config); err != nil {
		return nil, err
	}
	if conn == nil {
		return nil, errors.New("conn must not be nil")
	}
	ctx := context.Background()
	stmt, err := genPdbgStatement(ctx, conn, config.callTemplate, config.callArgs...)
	if err != nil {
		return nil, err
	}
	plsqlDebug := &PlsqlDebug{
		ctx:    ctx,
		Stmt:   stmt,
		config: config,
		conn:   conn,
	}
	if config.enableOutput {
		if err := initDebugOutputStmt(ctx, plsqlDebug); err != nil {
			_ = stmt.Close()
			return nil, err
		}
	}
	return plsqlDebug, nil
}

// NewPlsqlDebug 创建新的连接并初始化 PlsqlDebug（原有签名不变）。
func NewPlsqlDebug(dsn string, opts ...DebugConfigOpt) (*PlsqlDebug, error) {
	conn, err := GenYasconn(dsn)
	if err != nil {
		return nil, err
	}
	p, err := newPlsqlDebug(conn, opts...)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return p, nil
}

// NewPlsqlDebugWithConn 复用已有连接创建 PlsqlDebug。
// 调用方持有 conn 的所有权 — Close() 会关闭 conn；
// 用 CloseWithoutConn() 释放 PDBG 资源但保留共享连接。
func NewPlsqlDebugWithConn(conn *YasConn, opts ...DebugConfigOpt) (*PlsqlDebug, error) {
	return newPlsqlDebug(conn, opts...)
}

func validateDebugConfig(config *debugConfig) error {
	if len(config.callTemplate) <= 0 {
		return errors.New("call template must be provided")
	}
	return nil
}

func (p *PlsqlDebug) Start(objId uint64, subId uint16) error {
	return PdbgStart(p.Stmt, objId, subId)
}

func (p *PlsqlDebug) CheckVersion(objId uint64, subId uint16, version uint32) error {
	return PdbgCheckVersion(p.Stmt, objId, subId, version)
}

func (p *PlsqlDebug) Abort() error {
	return PdbgAbort(p.Stmt)
}

func (p *PlsqlDebug) Continue() error {
	return PdbgContinue(p.Stmt)
}

func (p *PlsqlDebug) StepInto() error {
	return PdbgStepInto(p.Stmt)
}

func (p *PlsqlDebug) StepOut() error {
	return PdbgStepOut(p.Stmt)
}

func (p *PlsqlDebug) StepNext() error {
	return PdbgStepNext(p.Stmt)
}

func (p *PlsqlDebug) DeleteAllBreakpoint() error {
	return PdbgDeleteAllBreakpoints(p.Stmt)
}

func (p *PlsqlDebug) AddBreakpoint(objId uint64, subId uint16, lineNo uint32) (uint32, error) {
	return PdbgAddBreakpoint(p.Stmt, objId, subId, lineNo)
}

func (p *PlsqlDebug) DeleteBreakpoint(objId uint64, subId uint16, lineNo uint32) error {
	return PdbgDeleteBreakpoint(p.Stmt, objId, subId, lineNo)
}

func (p *PlsqlDebug) GetBreakpointsCount() (uint32, error) {
	return PdbgGetBreakpointsCount(p.Stmt)
}

func (p *PlsqlDebug) GetAllVars() (uint32, error) {
	return PdbgGetAllVars(p.Stmt)
}

func (p *PlsqlDebug) GetAllFrames() (uint32, error) {
	return PdbgGetAllFrames(p.Stmt)
}

func (p *PlsqlDebug) GetRunningAttrs(attr DebugRunningAttr, value interface{}) error {
	return PdbgGetRunningAttrs(p.Stmt, attr, value)
}

func (p *PlsqlDebug) GetFrameAttrs(id uint32, attr DebugFrameAttr, value interface{}) error {
	return PdbgGetFrameAttrs(p.Stmt, id, attr, value)
}

func (p *PlsqlDebug) GetVarAttrs(id uint32, attr DebugVarAttr, value interface{}) error {
	return PdbgGetVarAttrs(p.Stmt, id, attr, value)
}

func (p *PlsqlDebug) GetVarValue(id uint32) (string, error) {
	return PdbgGetVarValue(p.Stmt, id)
}

func (p *PlsqlDebug) GetBreakpointAttrs(id uint32, attr DebugBpAttr, value interface{}) error {
	return PdbgGetBreakpointAttrs(p.Stmt, id, attr, value)
}

func (p *PlsqlDebug) GetAllRunningAttrs() (*PdbgRunningAttr, error) {
	return PdbgGetAllRunningAttrs(p.Stmt)
}

func (p *PlsqlDebug) GetAllBreakpointAttrs() ([]*PdbgBreakpointAttr, error) {
	return PdbgGetAllBreakpointAttrs(p.Stmt)
}

func (p *PlsqlDebug) GetAllVarAttrs() ([]*PdbgVarAttr, error) {
	return PdbgGetAllVarAttrs(p.Stmt)
}

func (p *PlsqlDebug) GetAllFrameAttrs() ([]*PdbgFrameData, error) {
	return PdbgGetAllFrameAttrs(p.Stmt)
}

// closeStmtsAndOutputBuf 释放输出缓冲和所有 PDBG 语句（私有）。
func (p *PlsqlDebug) closeStmtsAndOutputBuf() {
	if p.outputBuf != nil {
		C.free(unsafe.Pointer(p.outputBuf))
		p.outputBuf = nil
	}
	stmts := []*YasStmt{p.Stmt}
	if p.config.enableOutput {
		stmts = append(stmts, p.enableStmt, p.getLineStmt)
	}
	for _, stmt := range stmts {
		if stmt != nil {
			_ = PdbgStatementClose(stmt)
		}
	}
}

// Close 关闭所有语句并关闭底层连接（原有签名不变）。
func (p *PlsqlDebug) Close() error {
	p.closeStmtsAndOutputBuf()
	return p.conn.Close()
}

// CloseWithoutConn 关闭所有语句但保留底层连接。
// 用于共享连接场景：释放 PDBG 资源但不关闭 YasConn。
func (p *PlsqlDebug) CloseWithoutConn() error {
	p.closeStmtsAndOutputBuf()
	return nil
}

func (p *PlsqlDebug) GetBindOutValue() error {
	return PbdgGetBindOutValue(p.Stmt)
}

func (p *PlsqlDebug) PdbgGetOutput() (string, error) {
	if p.outputBuf == nil {
		p.outputBuf = (*C.char)(mallocBytes(PdbgOutputDefaultBufSize))
	}

	var sb strings.Builder
	for {
		length := C.uint32_t(PdbgOutputDefaultBufSize)
		var hasMore C.bool
		err := yapiPdbgGetOutput(p.Stmt.Stmt, p.outputBuf, &length, &hasMore)
		if err != nil {
			return sb.String(), err
		}
		if length > 0 {
			sb.WriteString(C.GoStringN(p.outputBuf, C.int(length)))
		}
		if !hasMore {
			break
		}
	}
	return sb.String(), nil
}

func (p *PlsqlDebug) GetOutput() ([]string, error) {
	var res []string
	if !p.config.enableOutput {
		return res, nil
	}
	if p.getLineStmt == nil {
		return nil, errors.New("output getLine statment is not init")
	}
	for {
		_, err := p.getLineStmt.exec()
		if err != nil {
			return nil, err
		}
		if p.getLineStatus != 0 {
			break
		}
		res = append(res, p.line)
	}
	return res, nil
}

func genPdbgStatement(ctx context.Context, conn *YasConn, sql string, args ...any) (*YasStmt, error) {
	stmt, err := PrepareContext(conn, ctx, sql)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if len(args) == 0 {
		return stmt, nil
	}
	nArgs, err := ConvertToNameValue(args...)
	if err != nil {
		return nil, err
	}
	if err := stmt.bindValues(nArgs); err != nil {
		_ = conn.Close()
		_ = stmt.Close()
		return nil, err
	}
	return stmt, nil
}

func initDebugOutputStmt(ctx context.Context, debug *PlsqlDebug) error {
	conn := debug.conn
	config := debug.config
	enableStmt, err := genPdbgStatement(ctx, conn, "BEGIN DBMS_OUTPUT.ENABLE(?); END;", config.outputCacheSize)
	if err != nil {
		return err
	}
	if _, err := enableStmt.exec(); err != nil {
		return err
	}
	debug.enableStmt = enableStmt
	lineOut, err := NewOutputBindValue(&debug.line, WithTypeVarchar(), WithBindSize(config.outputMaxLineLen))
	if err != nil {
		return err
	}
	lineStatusOut, err := NewOutputBindValue(&debug.getLineStatus, WithTypeInteger())
	if err != nil {
		return err
	}
	lineArgs := sql.Out{Dest: lineOut}
	getLineStatusArgs := sql.Out{Dest: lineStatusOut}
	getLineStmt, err := genPdbgStatement(ctx, conn, "BEGIN DBMS_OUTPUT.GET_LINE(?,?); END;", lineArgs, getLineStatusArgs)
	if err != nil {
		return err
	}
	debug.getLineStmt = getLineStmt
	return nil
}

func PdbgStart(stmt *YasStmt, objId uint64, subId uint16) error {
	return yapiPdbgStart(stmt.Stmt, C.uint64_t(objId), C.uint16_t(subId))
}

func PdbgCheckVersion(stmt *YasStmt, objId uint64, subId uint16, version uint32) error {
	return yapiPdbgCheckVersion(stmt.Stmt, C.uint64_t(objId), C.uint16_t(subId), C.uint32_t(version))
}

func PdbgAbort(stmt *YasStmt) error {
	return yapiPdbgAbort(stmt.Stmt)
}

func PdbgContinue(stmt *YasStmt) error {
	return yapiPdbgContinue(stmt.Stmt)
}

func PdbgStepInto(stmt *YasStmt) error {
	return yapiPdbgStepInto(stmt.Stmt)
}

func PdbgStepOut(stmt *YasStmt) error {
	return yapiPdbgStepOut(stmt.Stmt)
}

func PdbgStepNext(stmt *YasStmt) error {
	return yapiPdbgStepNext(stmt.Stmt)
}

func PbdgGetBindOutValue(stmt *YasStmt) error {
	if err := stmt.getBindValueDest(); err != nil {
		return err
	}
	return nil
}

func PdbgStatementClose(stmt *YasStmt) error {
	stmt.freeBindValues()
	if err := stmt.Close(); err != nil {
		return err
	}
	return nil
}

func PdbgDeleteAllBreakpoints(stmt *YasStmt) error {
	return yapiPdbgDeleteAllBreakpoints(stmt.Stmt)
}

func PdbgAddBreakpoint(stmt *YasStmt, objId uint64, subId uint16, lineNo uint32) (uint32, error) {
	bpID := new(C.uint32_t)
	err := yapiPdbgAddBreakpoint(stmt.Stmt, objId, subId, lineNo, bpID)
	if err != nil {
		return 0, err
	}
	return uint32(*bpID), nil
}

func PdbgDeleteBreakpoint(stmt *YasStmt, objId uint64, subId uint16, lineNo uint32) error {
	return yapiPdbgDeleteBreakpoint(stmt.Stmt, objId, subId, lineNo)
}

func PdbgGetBreakpointsCount(stmt *YasStmt) (uint32, error) {
	bpCount := new(C.uint32_t)
	err := yapiPdbgGetBreakpointsCount(stmt.Stmt, bpCount)
	if err != nil {
		return 0, err
	}
	return uint32(*bpCount), nil
}

func PdbgGetAllVars(stmt *YasStmt) (uint32, error) {
	varCount := new(C.uint32_t)
	err := yapiPdbgGetAllVars(stmt.Stmt, varCount)
	if err != nil {
		return 0, err
	}
	return uint32(*varCount), nil
}

func PdbgGetAllFrames(stmt *YasStmt) (uint32, error) {
	frameCount := new(C.uint32_t)
	err := yapiPdbgGetAllFrames(stmt.Stmt, frameCount)
	if err != nil {
		return 0, err
	}
	return uint32(*frameCount), nil
}

func PdbgGetRunningAttrs(stmt *YasStmt, attr DebugRunningAttr, value interface{}) error {
	switch attr {
	case DBG_RUNNING_ATTR_STATUS:
		data, ok := value.(*DebuggerStatus)
		if !ok {
			return fmt.Errorf("the value parameter type must be *DebuggerStatus")
		}
		status := new(C.uint32_t)
		stringLen := C.int32_t(0)
		err := yapiPdbgGetRunningAttrs(stmt.Stmt, C.YAPI_DBG_RUNNING_ATTR_STATUS, C.YapiPointer(status), 4, &stringLen)
		if err != nil {
			return err
		}
		if *status == C.uint32_t(C.YAPI_DBG_STATUS_ON) {
			*data = DBG_STATUS_ON
		} else {
			*data = DBG_STATUS_OFF
		}

	case DBG_RUNNING_ATTR_OBJ_ID:
		data, ok := value.(*uint64)
		if !ok {
			return ValueTypeUint64Err
		}

		objId := new(C.uint64_t)
		stringLen := C.int32_t(0)
		err := yapiPdbgGetRunningAttrs(stmt.Stmt, C.YAPI_DBG_RUNNING_ATTR_OBJ_ID, C.YapiPointer(objId), 8, &stringLen)
		if err != nil {
			return err
		}
		*data = uint64(*objId)

	case DBG_RUNNING_ATTR_SUB_ID:
		data, ok := value.(*uint16)
		if !ok {
			return ValueTypeUint16Err
		}

		subId := new(C.uint16_t)
		stringLen := C.int32_t(0)
		err := yapiPdbgGetRunningAttrs(stmt.Stmt, C.YAPI_DBG_RUNNING_ATTR_SUB_ID, C.YapiPointer(subId), 2, &stringLen)
		if err != nil {
			return err
		}
		*data = uint16(*subId)

	case DBG_RUNNING_ATTR_LINE_NO:
		data, ok := value.(*uint32)
		if !ok {
			return ValueTypeUint32Err
		}

		lineNo := new(C.uint32_t)
		stringLen := C.int32_t(0)
		err := yapiPdbgGetRunningAttrs(stmt.Stmt, C.YAPI_DBG_RUNNING_ATTR_LINE_NO, C.YapiPointer(lineNo), 4, &stringLen)
		if err != nil {
			return err
		}
		*data = uint32(*lineNo)

	case DBG_RUNNING_ATTR_CLASS_NAME:
		data, ok := value.(*string)
		if !ok {
			return ValueTypeStringErr
		}
		bufLen := DEFAULT_ATTRS_BUFFER * stmt.Conn.charsetRatio
		className := (*C.char)(mallocBytes(bufLen))
		stringLen := C.int32_t(0)
		defer C.free(unsafe.Pointer(className))
		err := yapiPdbgGetRunningAttrs(stmt.Stmt, C.YAPI_DBG_RUNNING_ATTR_CLASS_NAME, C.YapiPointer(className), C.int32_t(bufLen), &stringLen)
		if err != nil {
			return err
		}
		*data = C.GoString(className)

	case DBG_RUNNING_ATTR_METHOD_NAME:
		data, ok := value.(*string)
		if !ok {
			return ValueTypeStringErr
		}
		bufLen := DEFAULT_ATTRS_BUFFER * stmt.Conn.charsetRatio
		methodName := (*C.char)(mallocBytes(bufLen))
		stringLen := C.int32_t(0)
		defer C.free(unsafe.Pointer(methodName))
		err := yapiPdbgGetRunningAttrs(stmt.Stmt, C.YAPI_DBG_RUNNING_ATTR_METHOD_NAME, C.YapiPointer(methodName), C.int32_t(bufLen), &stringLen)
		if err != nil {
			return err
		}
		*data = C.GoString(methodName)

	default:
		return fmt.Errorf("unsupport debug running attr %v", attr)
	}
	return nil
}

func PdbgGetFrameAttrs(stmt *YasStmt, id uint32, attr DebugFrameAttr, value interface{}) error {
	switch attr {
	case DBG_FRAME_ATTR_OBJ_ID:
		data, ok := value.(*uint64)
		if !ok {
			return ValueTypeUint64Err
		}

		stringLen := C.int32_t(0)
		outValue := new(C.uint64_t)
		err := yapiPdbgGetFrameAttrs(stmt.Stmt, C.uint32_t(id), C.YAPI_DBG_FRAME_ATTR_OBJ_ID, C.YapiPointer(outValue), 8, &stringLen)
		if err != nil {
			return err
		}
		*data = uint64(*outValue)

	case DBG_FRAME_ATTR_SUB_ID:
		data, ok := value.(*uint16)
		if !ok {
			return ValueTypeUint16Err
		}
		stringLen := C.int32_t(0)
		outValue := new(C.uint16_t)
		err := yapiPdbgGetFrameAttrs(stmt.Stmt, C.uint32_t(id), C.YAPI_DBG_FRAME_ATTR_SUB_ID, C.YapiPointer(outValue), 2, &stringLen)
		if err != nil {
			return err
		}
		*data = uint16(*outValue)

	case DBG_FRAME_ATTR_LINE_NO:
		data, ok := value.(*uint32)
		if !ok {
			return ValueTypeUint32Err
		}
		stringLen := C.int32_t(0)
		outValue := new(C.uint32_t)
		err := yapiPdbgGetFrameAttrs(stmt.Stmt, C.uint32_t(id), C.YAPI_DBG_FRAME_ATTR_LINE_NO, C.YapiPointer(outValue), 4, &stringLen)
		if err != nil {
			return err
		}
		*data = uint32(*outValue)

	case DBG_FRAME_ATTR_STACK_NO:
		data, ok := value.(*uint32)
		if !ok {
			return ValueTypeUint32Err
		}

		stringLen := C.int32_t(0)
		outValue := new(C.uint32_t)
		err := yapiPdbgGetFrameAttrs(stmt.Stmt, C.uint32_t(id), C.YAPI_DBG_FRAME_ATTR_STACK_NO, C.YapiPointer(outValue), 4, &stringLen)
		if err != nil {
			return err
		}
		*data = uint32(*outValue)

	case DBG_FRAME_ATTR_CLASS_NAME:
		data, ok := value.(*string)
		if !ok {
			return ValueTypeStringErr
		}
		stringLen := C.int32_t(0)
		bufLen := DEFAULT_ATTRS_BUFFER * stmt.Conn.charsetRatio
		outValue := (*C.char)(mallocBytes(bufLen))
		defer C.free(unsafe.Pointer(outValue))

		err := yapiPdbgGetFrameAttrs(stmt.Stmt, C.uint32_t(id), C.YAPI_DBG_FRAME_ATTR_CLASS_NAME, C.YapiPointer(outValue), C.int32_t(bufLen), &stringLen)
		if err != nil {
			return err
		}
		*data = C.GoString(outValue)

	case DBG_FRAME_ATTR_METHOD_NAME:
		data, ok := value.(*string)
		if !ok {
			return ValueTypeStringErr
		}

		stringLen := C.int32_t(0)
		bufLen := DEFAULT_ATTRS_BUFFER * stmt.Conn.charsetRatio
		outValue := (*C.char)(mallocBytes(bufLen))
		defer C.free(unsafe.Pointer(outValue))
		err := yapiPdbgGetFrameAttrs(stmt.Stmt, C.uint32_t(id), C.YAPI_DBG_FRAME_ATTR_METHOD_NAME, C.YapiPointer(outValue), C.int32_t(bufLen), &stringLen)
		if err != nil {
			return err
		}
		*data = C.GoString(outValue)

	default:
		return fmt.Errorf("unsupport debug frame attr %v", attr)
	}
	return nil
}

func PdbgGetVarAttrs(stmt *YasStmt, id uint32, attr DebugVarAttr, value interface{}) error {
	switch attr {
	case DBG_VAR_ATTR_BLOCK_NO:
		data, ok := value.(*uint32)
		if !ok {
			return ValueTypeUint32Err
		}

		stringLen := C.int32_t(0)
		outValue := new(C.uint32_t)
		err := yapiPdbgGetVarAttrs(stmt.Stmt, C.uint32_t(id), C.YAPI_DBG_VAR_ATTR_BLOCK_NO, C.YapiPointer(outValue), 4, &stringLen)
		if err != nil {
			return err
		}
		*data = uint32(*outValue)

	case DBG_VAR_ATTR_TYPE:
		data, ok := value.(*uint8)
		if !ok {
			return fmt.Errorf("the value parameter type must be *uint8")
		}

		stringLen := C.int32_t(0)
		outValue := new(C.uint8_t)
		err := yapiPdbgGetVarAttrs(stmt.Stmt, C.uint32_t(id), C.YAPI_DBG_VAR_ATTR_TYPE, C.YapiPointer(outValue), 1, &stringLen)
		if err != nil {
			return err
		}
		*data = uint8(*outValue)

	case DBG_VAR_ATTR_IS_GLOBAL:
		data, ok := value.(*uint8)
		if !ok {
			return fmt.Errorf("the value parameter type must be *uint8")
		}

		stringLen := C.int32_t(0)
		outValue := new(C.uint8_t)
		err := yapiPdbgGetVarAttrs(stmt.Stmt, C.uint32_t(id), C.YAPI_DBG_VAR_ATTR_IS_GLOBAL, C.YapiPointer(outValue), 1, &stringLen)
		if err != nil {
			return err
		}
		*data = uint8(*outValue)

	case DBG_VAR_ATTR_NAME:
		data, ok := value.(*string)
		if !ok {
			return ValueTypeStringErr
		}

		stringLen := C.int32_t(0)
		bufLen := DEFAULT_ATTRS_BUFFER * stmt.Conn.charsetRatio
		outValue := (*C.char)(mallocBytes(bufLen))
		defer C.free(unsafe.Pointer(outValue))

		err := yapiPdbgGetVarAttrs(stmt.Stmt, C.uint32_t(id), C.YAPI_DBG_VAR_ATTR_NAME, C.YapiPointer(outValue), C.int32_t(bufLen), &stringLen)
		if err != nil {
			return err
		}
		*data = C.GoString(outValue)

	case DBG_VAR_ATTR_VALUE_SIZE:
		data, ok := value.(*uint32)
		if !ok {
			return ValueTypeUint32Err
		}
		stringLen := C.int32_t(0)
		outValue := new(C.uint32_t)
		err := yapiPdbgGetVarAttrs(stmt.Stmt, C.uint32_t(id), C.YAPI_DBG_VAR_ATTR_VALUE_SIZE, C.YapiPointer(outValue), 4, &stringLen)
		if err != nil {
			return err
		}
		*data = uint32(*outValue)

	default:
		return fmt.Errorf("unsupport debug var attr %v", attr)
	}

	return nil
}

func PdbgGetVarValue(stmt *YasStmt, id uint32) (string, error) {
	var (
		bufLen     int32
		value      C.YapiPointer
		bindType   C.YapiType
		actualType C.YapiType
		isLob      bool
		err        error
		dataType   uint8
		valueSize  uint32
	)
	if err := PdbgGetVarAttrs(stmt, id, DBG_VAR_ATTR_TYPE, &dataType); err != nil {
		return "", err
	}

	if err := PdbgGetVarAttrs(stmt, id, DBG_VAR_ATTR_VALUE_SIZE, &valueSize); err != nil {
		return "", err
	}

	bindType = C.YAPI_TYPE_VARCHAR
	actualType = C.YapiType(dataType)

	switch actualType {
	case C.YAPI_TYPE_NCHAR, C.YAPI_TYPE_NVARCHAR:
		bufLen = int32(stmt.Conn.ncharsetRatio*valueSize) + 1
		value = C.YapiPointer(mallocBytes(uint32(bufLen)))
	case C.YAPI_TYPE_CHAR, C.YAPI_TYPE_VARCHAR:
		bufLen = int32(stmt.Conn.charsetRatio*valueSize) + 1
		value = C.YapiPointer(mallocBytes(uint32(bufLen)))
	case C.YAPI_TYPE_BIT:
		bufLen = int32(valueSize*8) + 1
		value = C.YapiPointer(mallocBytes(uint32(bufLen)))
	case C.YAPI_TYPE_BINARY:
		bufLen = int32(sizeToAlign4(valueSize*2)) + 1
		value = C.YapiPointer(mallocBytes(uint32(bufLen)))

	case C.YAPI_TYPE_CLOB, C.YAPI_TYPE_BLOB:
		desc, err := stmt.Conn.lobWrite(actualType, nil)
		if err != nil {
			return "", err
		}
		value = C.YapiPointer(desc)
		bufLen = -1
		isLob = true
		bindType = actualType
	default:
		bufLen = 125
		value = C.YapiPointer(mallocBytes(uint32(bufLen)))
	}

	indicator := new(C.int32_t)
	err = yapiPdbgGetVarValue(stmt.Stmt, C.uint32_t(id), C.uint32_t(bindType), value, C.int32_t(bufLen), indicator)
	if err != nil {
		return "", err
	}

	defer func() {
		if isLob {
			lobLocator := (**C.YapiLobLocator)(unsafe.Pointer(value))
			stmt.Conn.lobFree(actualType, *lobLocator)
		} else {
			C.free(unsafe.Pointer(value))
		}
	}()

	if *indicator == C.YAPI_NULL_DATA {
		return "", nil
	}

	if isLob {
		lobLocator := (**C.YapiLobLocator)(value)
		byteValue, err := stmt.Conn.lobRead(*lobLocator)
		if err != nil {
			return "", err
		}
		return string(byteValue), nil
	}
	rawStr := C.GoString((*C.char)(value))

	switch actualType {
	case C.YAPI_TYPE_DATE, C.YAPI_TYPE_SHORTDATE, C.YAPI_TYPE_SHORTTIME, C.YAPI_TYPE_TIMESTAMP, C.YAPI_TYPE_TIMESTAMP_TZ, C.YAPI_TYPE_TIMESTAMP_LTZ:
		rawStr = fixTimeString(rawStr)
	}
	return rawStr, nil
}

func PdbgGetBreakpointAttrs(stmt *YasStmt, id uint32, attr DebugBpAttr, value interface{}) error {
	switch attr {
	case DBG_BP_ATTR_OBJ_ID:
		data, ok := value.(*uint64)
		if !ok {
			return ValueTypeUint64Err
		}

		stringLen := C.int32_t(0)
		outValue := new(C.uint64_t)
		err := yapiPdbgGetBreakpointAttrs(stmt.Stmt, C.uint32_t(id), C.YAPI_DBG_BP_ATTR_OBJ_ID, C.YapiPointer(outValue), 8, &stringLen)
		if err != nil {
			return err
		}
		*data = uint64(*outValue)

	case DBG_BP_ATTR_SUB_ID:
		data, ok := value.(*uint16)
		if !ok {
			return ValueTypeUint16Err
		}

		stringLen := C.int32_t(0)
		outValue := new(C.uint16_t)
		err := yapiPdbgGetBreakpointAttrs(stmt.Stmt, C.uint32_t(id), C.YAPI_DBG_BP_ATTR_SUB_ID, C.YapiPointer(outValue), 2, &stringLen)
		if err != nil {
			return err
		}
		*data = uint16(*outValue)

	case DBG_BP_ATTR_LINE_NO:
		data, ok := value.(*uint32)
		if !ok {
			return ValueTypeUint32Err
		}
		stringLen := C.int32_t(0)
		outValue := new(C.uint32_t)
		err := yapiPdbgGetBreakpointAttrs(stmt.Stmt, C.uint32_t(id), C.YAPI_DBG_BP_ATTR_LINE_NO, C.YapiPointer(outValue), 4, &stringLen)
		if err != nil {
			return err
		}
		*data = uint32(*outValue)

	default:
		return fmt.Errorf("unsupport debug breakpoint attr %v", attr)
	}

	return nil
}

type PdbgRunningAttr struct {
	Status     DebuggerStatus
	ObjId      uint64
	SubId      uint16
	LineNo     uint32
	ClassName  string
	MethodName string
}

func PdbgGetAllRunningAttrs(stmt *YasStmt) (*PdbgRunningAttr, error) {
	data := PdbgRunningAttr{}
	if err := PdbgGetRunningAttrs(stmt, DBG_RUNNING_ATTR_STATUS, &data.Status); err != nil {
		return nil, err
	}

	if err := PdbgGetRunningAttrs(stmt, DBG_RUNNING_ATTR_OBJ_ID, &data.ObjId); err != nil {
		return nil, err
	}

	if err := PdbgGetRunningAttrs(stmt, DBG_RUNNING_ATTR_SUB_ID, &data.SubId); err != nil {
		return nil, err
	}

	if err := PdbgGetRunningAttrs(stmt, DBG_RUNNING_ATTR_CLASS_NAME, &data.ClassName); err != nil {
		return nil, err
	}

	if err := PdbgGetRunningAttrs(stmt, DBG_RUNNING_ATTR_METHOD_NAME, &data.MethodName); err != nil {
		return nil, err
	}

	if err := PdbgGetRunningAttrs(stmt, DBG_RUNNING_ATTR_LINE_NO, &data.LineNo); err != nil {
		return nil, err
	}

	return &data, nil
}

type PdbgBreakpointAttr struct {
	Id     uint32
	ObjId  uint64
	SubId  uint16
	LineNo uint32
}

func PdbgGetAllBreakpointAttrs(stmt *YasStmt) ([]*PdbgBreakpointAttr, error) {
	bpCount, err := PdbgGetBreakpointsCount(stmt)
	if err != nil {
		return nil, err
	}

	bpDatas := make([]*PdbgBreakpointAttr, 0, bpCount)
	for i := uint32(0); i < bpCount; i++ {
		data := PdbgBreakpointAttr{Id: i}
		if err := PdbgGetBreakpointAttrs(stmt, i, DBG_BP_ATTR_OBJ_ID, &data.ObjId); err != nil {
			return nil, err
		}

		if err := PdbgGetBreakpointAttrs(stmt, i, DBG_BP_ATTR_SUB_ID, &data.SubId); err != nil {
			return nil, err
		}

		if err := PdbgGetBreakpointAttrs(stmt, i, DBG_BP_ATTR_LINE_NO, &data.LineNo); err != nil {
			return nil, err
		}
		bpDatas = append(bpDatas, &data)
	}

	return bpDatas, nil
}

type PdbgVarAttr struct {
	Id           uint32
	BlockNo      uint32
	DataType     uint8
	DataTypeName string
	GlobalLevel  uint8
	Name         string
	Size         uint32
	Value        string
}

func PdbgGetAllVarAttrs(stmt *YasStmt) ([]*PdbgVarAttr, error) {
	varCount, err := PdbgGetAllVars(stmt)
	if err != nil {
		return nil, err
	}
	varDatas := make([]*PdbgVarAttr, 0, varCount)
	for i := uint32(0); i < varCount; i++ {
		data := PdbgVarAttr{Id: i}
		if err := PdbgGetVarAttrs(stmt, i, DBG_VAR_ATTR_BLOCK_NO, &data.BlockNo); err != nil {
			return nil, err
		}

		if err := PdbgGetVarAttrs(stmt, i, DBG_VAR_ATTR_TYPE, &data.DataType); err != nil {
			return nil, err
		}

		data.DataTypeName = GetDatabaseTypeName(uint32(data.DataType))

		if err := PdbgGetVarAttrs(stmt, i, DBG_VAR_ATTR_IS_GLOBAL, &data.GlobalLevel); err != nil {
			return nil, err
		}

		if err := PdbgGetVarAttrs(stmt, i, DBG_VAR_ATTR_NAME, &data.Name); err != nil {
			return nil, err
		}

		if err := PdbgGetVarAttrs(stmt, i, DBG_VAR_ATTR_VALUE_SIZE, &data.Size); err != nil {
			return nil, err
		}

		value, err := PdbgGetVarValue(stmt, i)
		if err != nil {
			return nil, err
		}
		data.Value = value
		varDatas = append(varDatas, &data)
	}
	return varDatas, nil
}

type PdbgFrameData struct {
	Id         uint32
	ObjId      uint64
	SubId      uint16
	LineNo     uint32
	StackNo    uint32
	ClassName  string
	MethodName string
}

func PdbgGetAllFrameAttrs(stmt *YasStmt) ([]*PdbgFrameData, error) {
	frameCount, err := PdbgGetAllFrames(stmt)
	if err != nil {
		return nil, err
	}
	frameDatas := make([]*PdbgFrameData, 0, frameCount)
	for i := uint32(0); i < frameCount; i++ {
		data := PdbgFrameData{Id: i}
		if err := PdbgGetFrameAttrs(stmt, i, DBG_FRAME_ATTR_OBJ_ID, &data.ObjId); err != nil {
			return nil, err
		}

		if err := PdbgGetFrameAttrs(stmt, i, DBG_FRAME_ATTR_SUB_ID, &data.SubId); err != nil {
			return nil, err
		}

		if err := PdbgGetFrameAttrs(stmt, i, DBG_FRAME_ATTR_LINE_NO, &data.LineNo); err != nil {
			return nil, err
		}

		if err := PdbgGetFrameAttrs(stmt, i, DBG_FRAME_ATTR_STACK_NO, &data.StackNo); err != nil {
			return nil, err
		}

		if err := PdbgGetFrameAttrs(stmt, i, DBG_FRAME_ATTR_CLASS_NAME, &data.ClassName); err != nil {
			return nil, err
		}

		if err := PdbgGetFrameAttrs(stmt, i, DBG_FRAME_ATTR_METHOD_NAME, &data.MethodName); err != nil {
			return nil, err
		}

		frameDatas = append(frameDatas, &data)
	}
	return frameDatas, nil
}

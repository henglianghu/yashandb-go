package yasdb

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

var (
	procName_1 = `PROCADD`
	plsql_1    = `create or replace procedure procAdd(p1 in int, p2 in int, p3 out int) 
	           is 
	           begin
               p3 := 100; 
			   p3 := p1 + p2;
               end;`

	callPlSql_1 = `call procAdd(?,?,?)`

	procName_2 = "LX_PROC"
	plsql_2    = `
	CREATE OR REPLACE PROCEDURE LX_PROC(i INT,b varchar) is
	begin
		DBMS_OUTPUT.PUT_LINE(to_char(i));
		DBMS_OUTPUT.PUT_LINE(b);
	end;`
	callPlSql_2 = `call LX_PROC(?,?)`

	procName_3 = `FUNC_OUTPARAM`
	plsql_3    = `CREATE OR REPLACE function func_outparam(c1 out int,c2 out float,c3 out double,c4 out varchar,c5 out char,c6 out date,c7 out boolean,c8 out clob,c9 out rowid,c10 out json,c11 out nchar,c12 out nvarchar) return varchar is
	res varchar(8000);
	v1 int := 943093745;
	v2 float := 1506141.9;
	v3 double := 107175737.7;
	v4 varchar(20) := 'yasdb';
	v5 char(10) := 'yasql';
	v6 date := '2023-01-20';
	v7 boolean := false;
	v8 clob := 'It gives me great pleasure to introduce our company.';
	v9 rowid := '1350:5:0:148:0';
	v10 json := '{"name":"Jack", "city":"Beijing","school":"TsingHua University"}';
	v11 nchar(10) := '😂😂😼😶😶';
	v12 nvarchar(13) := '中国深圳市龙华区崖山数据库';
	begin
	c1 := v1;
	c2 := v2;
	c3 := v3;
	c4 := v4;
	c5 := v5;
	c6 := v6;
	c7 := v7;
	c8 := v8;
	c9 := v9;
	c10 := v10;
	c11 := v11;
	c12 := v12;
	res := c1||':'||c4||':'||c5||':'||c6||':'||c7||':'||c8||':'||c9 || ':' || c10 || ':' || c11 || ':' || c12;
	return res;
	end;`

	callPlSql_3 = `DECLARE
	v_result VARCHAR(8000);
  C1 INTEGER;
  C2 FLOAT;
  C3 DOUBLE;
  C4 VARCHAR(8000);
  C5 CHAR(1000);
  C6 DATE;
  C7 BOOLEAN;
  C8 CLOB;
  C9 ROWID;
  C10 JSON;
  c11 nchar(10);
  c12 nvarchar(13);
  BEGIN
	v_result := FUNC_OUTPARAM(C1, C2, C3, C4, C5, C6, C7, C8, C9, C10, c11, c12);
  END;`
)

func createProcedute(t *testing.T, sqlStr string) {
	db, err := sql.Open("yasdb", fmt.Sprintf("%s?%s", testDsn, "autocommit=true"))
	if err != nil {
		t.Fatalf("open database err: %v", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(sqlStr)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
}

func queryObjIdAndSubId(t *testing.T, proceName string) (uint64, uint16) {
	db, err := sql.Open("yasdb", fmt.Sprintf("%s?%s", testDsn, "autocommit=true"))
	if err != nil {
		t.Fatalf("open database err: %v", err)
		return 0, 0
	}
	defer db.Close()

	querySql := `select OBJECT_ID,SUBPROGRAM_ID from dba_procedures where  object_name = ?;`
	rows, err := db.Query(querySql, proceName)
	if err != nil {
		t.Fatalf("exec %s failed, %s", querySql, err.Error())
	}
	defer rows.Close()
	var (
		objId uint64
		subId uint16
	)
	for rows.Next() {
		err := rows.Scan(&objId, &subId)
		if err != nil {
			t.Fatalf(err.Error())
		}
	}
	return objId, subId
}

func TestPdbgStart(t *testing.T) {
	createProcedute(t, plsql_1)
	out := 0
	v1 := 1
	v2 := 100
	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(callPlSql_1, v1, v2, sql.Out{Dest: &out}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}
	if err := p.Abort(); err != nil {
		t.Fatal(err)
	}
}

func TestPdbgContinte(t *testing.T) {
	createProcedute(t, plsql_1)
	out := 0
	v1 := 102
	v2 := 100
	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(callPlSql_1, v1, v2, sql.Out{Dest: &out}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}
	if err := p.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := p.GetBindOutValue(); err != nil {
		t.Fatal(err)
	}
	if out != (v1 + v2) {
		t.Fatalf("bind out value %d != %d", out, v1+v2)
	}
	fmt.Println(out)
}

func TestPdgStepNextStepInto(t *testing.T) {
	createProcedute(t, plsql_1)
	out := int64(0)
	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(callPlSql_1, int64(1), int64(2), sql.Out{Dest: &out}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}
	if err := p.StepInto(); err != nil {
		t.Fatal(err)
	}
	if err := p.StepInto(); err != nil {
		t.Fatal(err)
	}
	if err := p.StepInto(); err != nil {
		t.Fatal(err)
	}
}

func TestPdgStepNext(t *testing.T) {
	createProcedute(t, plsql_1)
	out := int64(0)
	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(callPlSql_1, int64(1), int64(2), sql.Out{Dest: &out}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}
	if err := p.StepNext(); err != nil {
		t.Fatal(err)
	}
	if err := p.Abort(); err != nil {
		t.Fatal(err)
	}
}

func TestPdbgStepOut(t *testing.T) {
	createProcedute(t, plsql_1)
	out := int64(0)
	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(callPlSql_1, int64(1), int64(2), sql.Out{Dest: &out}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}
	if err := p.StepOut(); err != nil {
		t.Fatal(err)
	}
}

func TestPdbgGetRunningAttrs(t *testing.T) {
	createProcedute(t, plsql_1)
	out := 0
	v1 := 102
	v2 := 100
	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(callPlSql_1, v1, v2, sql.Out{Dest: &out}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}

	var status DebuggerStatus
	if err := PdbgGetRunningAttrs(p.Stmt, DBG_RUNNING_ATTR_STATUS, &status); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("running debugger status: %v\n", status)

	var runningObjId uint64
	if err := PdbgGetRunningAttrs(p.Stmt, DBG_RUNNING_ATTR_OBJ_ID, &runningObjId); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("running obj id: %v\n", runningObjId)

	var className string
	if err := PdbgGetRunningAttrs(p.Stmt, DBG_RUNNING_ATTR_CLASS_NAME, &className); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("running class name: %v\n", className)

	var methodName string
	if err := PdbgGetRunningAttrs(p.Stmt, DBG_RUNNING_ATTR_METHOD_NAME, &methodName); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("running method name: %v\n", methodName)

	if err := p.Continue(); err != nil {
		t.Fatal(err)
	}

	if err := PdbgGetRunningAttrs(p.Stmt, DBG_RUNNING_ATTR_STATUS, &status); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("running debugger status: %v\n", status)

	if err := p.GetBindOutValue(); err != nil {
		t.Fatal(err)
	}
	if out != (v1 + v2) {
		t.Fatalf("bind out value %d != %d", out, v1+v2)
	}
	fmt.Println(out)
}

func TestPdbgGetFrameAttrs(t *testing.T) {
	createProcedute(t, plsql_1)
	out := 0
	v1 := 102
	v2 := 100
	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(callPlSql_1, v1, v2, sql.Out{Dest: &out}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}

	if err := PdbgStepInto(p.Stmt); err != nil {
		t.Fatal(err)
	}

	count, err := PdbgGetAllFrames(p.Stmt)
	if err != nil {
		t.Fatal(err)
	}

	for i := uint32(0); i < count; i++ {
		var objId uint64
		if err := PdbgGetFrameAttrs(p.Stmt, i, DBG_FRAME_ATTR_OBJ_ID, &objId); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("frame: %d ,object id: %d\n", i, objId)

		var subId uint16
		if err := PdbgGetFrameAttrs(p.Stmt, i, DBG_FRAME_ATTR_SUB_ID, &subId); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("frame: %d ,subid: %d\n", i, subId)

		var lineNo uint32
		if err := PdbgGetFrameAttrs(p.Stmt, i, DBG_FRAME_ATTR_LINE_NO, &lineNo); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("frame: %d ,lineNo: %d\n", i, subId)

		var stackNo uint32
		if err := PdbgGetFrameAttrs(p.Stmt, i, DBG_FRAME_ATTR_STACK_NO, &stackNo); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("frame: %d ,stackNo: %d\n", i, subId)

		var className string
		if err := PdbgGetFrameAttrs(p.Stmt, i, DBG_FRAME_ATTR_CLASS_NAME, &className); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("frame: %d ,className: %s\n", i, className)

		var methodName string
		if err := PdbgGetFrameAttrs(p.Stmt, i, DBG_FRAME_ATTR_METHOD_NAME, &methodName); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("frame: %d ,methodName: %s\n", i, methodName)

	}

	if err := p.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := p.GetBindOutValue(); err != nil {
		t.Fatal(err)
	}
	if out != (v1 + v2) {
		t.Fatalf("bind out value %d != %d", out, v1+v2)
	}
	fmt.Println(out)
}

func TestPdbgGetAllVars(t *testing.T) {
	createProcedute(t, plsql_1)
	out := 0
	v1 := 102
	v2 := 100
	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(callPlSql_1, v1, v2, sql.Out{Dest: &out}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}
	varCount, err := p.GetAllVars()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("values count is %v\n", varCount)
}

func TestPdbgGetAllFrames(t *testing.T) {
	createProcedute(t, plsql_1)
	out := 0
	v1 := 102
	v2 := 100
	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(callPlSql_1, v1, v2, sql.Out{Dest: &out}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}
	varCount, err := p.GetAllFrames()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("frame count is %v\n", varCount)
}

func TestPdbgAddBreakpoint(t *testing.T) {
	createProcedute(t, plsql_1)
	out := int64(0)
	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(callPlSql_1, int64(1), int64(2), sql.Out{Dest: &out}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}

	bpId, err := p.AddBreakpoint(objId, subId, 4)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("breakpoint id is %v\n", bpId)

	bpId, err = p.AddBreakpoint(objId, subId, 5)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("breakpoint id is %v\n", bpId)

	if err := p.Continue(); err != nil {
		t.Fatal(err)
	}

	// if err := p.Continue(); err != nil {
	// 	t.Fatal(err)
	// }
}

func TestPdbgGetBreakpointAttrs(t *testing.T) {
	createProcedute(t, plsql_1)
	out := int64(0)
	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(callPlSql_1, int64(1), int64(2), sql.Out{Dest: &out}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}

	bpId, err := p.AddBreakpoint(objId, subId, 4)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("breakpoint id is %v\n", bpId)

	bpId, err = p.AddBreakpoint(2361, 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("breakpoint id is %v\n", bpId)

	bpCount, err := p.GetBreakpointsCount()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("breakpoint count is %v\n", bpCount)

	for i := uint32(0); i < bpCount; i++ {
		var objId uint64
		if err := p.GetBreakpointAttrs(i, DBG_BP_ATTR_OBJ_ID, &objId); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("breakpoit %d objectId is %v\n", i, objId)

		var subId uint16
		if err := p.GetBreakpointAttrs(i, DBG_BP_ATTR_SUB_ID, &subId); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("breakpoit %d subId is %v\n", i, subId)

		var lineNo uint32
		if err := p.GetBreakpointAttrs(i, DBG_BP_ATTR_LINE_NO, &lineNo); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("breakpoit %d lineNo is %v\n", i, lineNo)
	}
}

func TestPdbgDeleteBrakPoint(t *testing.T) {
	createProcedute(t, plsql_1)
	out := int64(0)
	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(callPlSql_1, int64(1), int64(2), sql.Out{Dest: &out}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}

	bpId, err := p.AddBreakpoint(objId, subId, 4)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("breakpoint id is %v\n", bpId)

	bpId, err = p.AddBreakpoint(objId, subId, 5)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("breakpoint id is %v\n", bpId)

	bpCount, err := p.GetBreakpointsCount()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("breakpoint count is %v\n", bpCount)

	i := uint32(1)
	var bpObjId uint64
	if err := p.GetBreakpointAttrs(i, DBG_BP_ATTR_OBJ_ID, &bpObjId); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("breakpoit %d objectId is %v\n", i, bpObjId)

	var bpSubId uint16
	if err := p.GetBreakpointAttrs(i, DBG_BP_ATTR_SUB_ID, &bpSubId); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("breakpoit %d subId is %v\n", i, subId)

	var lineNo uint32
	if err := p.GetBreakpointAttrs(i, DBG_BP_ATTR_LINE_NO, &lineNo); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("breakpoit %d lineNo is %v\n", i, lineNo)

	if err := p.DeleteBreakpoint(objId, subId, lineNo); err != nil {
		t.Fatal(err)
	}

	bpCount, err = p.GetBreakpointsCount()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("breakpoint count is %v\n", bpCount)
}

func TestPdbgDeleteAllBreakpoints(t *testing.T) {
	createProcedute(t, plsql_1)
	out := int64(0)
	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(callPlSql_1, int64(1), int64(2), sql.Out{Dest: &out}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}

	bpId, err := p.AddBreakpoint(objId, subId, 4)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("breakpoint id is %v\n", bpId)

	bpId, err = p.AddBreakpoint(objId, subId, 5)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("breakpoint id is %v\n", bpId)

	bpCount, err := p.GetBreakpointsCount()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("after delete all breakpoint count is %v\n", bpCount)

	if err := p.DeleteAllBreakpoint(); err != nil {
		t.Fatal(err)
	}

	bpCount, err = p.GetBreakpointsCount()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("breakpoint count is %v\n", bpCount)
}

func TestPdbgGetVarAttrs(t *testing.T) {
	createProcedute(t, plsql_1)
	out := 0
	v1 := 102
	v2 := 100
	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(callPlSql_1, v1, v2, sql.Out{Dest: &out}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}

	varCount, err := p.GetAllVars()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("values count is %v\n", varCount)

	for i := uint32(0); i < varCount; i++ {
		var blockNo uint32
		if err := p.GetVarAttrs(i, DBG_VAR_ATTR_BLOCK_NO, &blockNo); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("var %d blockNo is %v\n", i, blockNo)

		var dataType uint8
		if err := p.GetVarAttrs(i, DBG_VAR_ATTR_TYPE, &dataType); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("var %d dataType number is %v; dataType name is %s\n", i, dataType, GetDatabaseTypeName(uint32(dataType)))

		var globalLevel uint8
		if err := p.GetVarAttrs(i, DBG_VAR_ATTR_IS_GLOBAL, &globalLevel); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("var %d global level is %v\n", i, globalLevel)

		var name string
		if err := p.GetVarAttrs(i, DBG_VAR_ATTR_NAME, &name); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("var %d name is %v\n", i, name)

		var valueSize uint32
		if err := p.GetVarAttrs(i, DBG_VAR_ATTR_VALUE_SIZE, &valueSize); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("var %d valueSize is %v\n", i, valueSize)

	}
}

func TestPdbgGetVarValue(t *testing.T) {
	createProcedute(t, plsql_1)
	out := 0
	v1 := 102
	v2 := 100
	expected := 202
	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(callPlSql_1, v1, v2, sql.Out{Dest: &out}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}
	varCount, err := p.GetAllVars()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("values count is %v\n", varCount)

	for i := uint32(0); i < varCount; i++ {
		value, err := PdbgGetVarValue(p.Stmt, i)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("var %d value is %v\n", i, value)

	}

	if err := p.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := p.GetBindOutValue(); err != nil {
		t.Fatal(err)
	}
	if out != expected {
		t.Fatalf("out: %d expected: %d", out, expected)
	}

}

func TestPdbgNameBinding(t *testing.T) {

	add := `create or replace procedure procAdd(p1 in int, p2 in out int, p3 out int) 
	           is 
	           begin
               p3 := 100; 
			   p3 := p1 + p2;
			   p2 := 1;
               end;`
	call := `
	begin
		procAdd(
			p1 =>:p1,
			p2 =>:p2,
			p3 =>:p3
		);
	end;
	`

	createProcedute(t, add)
	p1 := 102
	p2 := 100
	p3 := 0
	p3Expected := 202
	p2Expected := 1

	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(call,
		sql.NamedArg{Name: "p3", Value: sql.Out{Dest: &p3}},
		sql.NamedArg{Name: "p1", Value: p1},
		sql.NamedArg{Name: "p2", Value: sql.Out{Dest: &p2, In: true}}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}
	if err = p.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := p.GetBindOutValue(); err != nil {
		t.Fatal(err)
	}

	if p3 != p3Expected {
		t.Fatalf("p3: %d p3Expected: %d", p3, p3Expected)
	}
	if p2 != p2Expected {
		t.Fatalf("p2: %d p3Expected: %d", p2, p2Expected)
	}

}

func TestPdbgNameReturnBinding(t *testing.T) {

	add := `
	CREATE OR REPLACE FUNCTION calculate_sum (
		a IN NUMBER,
		b IN NUMBER
	) RETURN NUMBER
	IS
		result NUMBER;
	BEGIN
		result := a + b;
		RETURN result;
	END;
	`
	call := `
	begin
		:result := calculate_sum(
			a =>:a,
			b =>:b
		);
	end;
	`

	createProcedute(t, add)
	a := float64(1)
	b := float64(1)
	result := float64(0)
	expected := float64(2)

	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(call,
		sql.NamedArg{Name: "result", Value: sql.Out{Dest: &result}},
		sql.NamedArg{Name: "a", Value: a},
		sql.NamedArg{Name: "b", Value: b}))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	objId, subId := queryObjIdAndSubId(t, procName_1)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}
	if err = p.Continue(); err != nil {
		t.Fatal(err)
	}
	if err := p.GetBindOutValue(); err != nil {
		t.Fatal(err)
	}

	if result != expected {
		t.Fatalf("result: %f expected: %f", result, expected)
	}

}

func TestPdbgGetAllData(t *testing.T) {
	createProcedute(t, plsql_3)
	p, err := NewPlsqlDebug(testDsn, WithDebugCallTempalate(callPlSql_3))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	objId, subId := queryObjIdAndSubId(t, procName_3)
	if err := p.Start(objId, subId); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 11; i++ {
		fmt.Println("step: ", i)
		StepNextAndprintVar(p, t)
	}
}

func StepNextAndprintVar(p *PlsqlDebug, t *testing.T) {
	runningData, err := p.GetAllRunningAttrs()
	if err != nil {
		t.Fatal(err)
	}
	printRuningData(runningData)

	if err := p.StepNext(); err != nil {
		t.Fatalf("step next failed， %v", err)
	}

	runningData, err = p.GetAllRunningAttrs()
	if err != nil {
		t.Fatal(err)
	}
	printRuningData(runningData)

	varDatas, err := p.GetAllVarAttrs()
	if err != nil {
		t.Fatal(err)
	}
	printVarData(varDatas)
	fmt.Println("=====================================================================================")
}

func printRuningData(data *PdbgRunningAttr) {
	fmt.Printf("LineNo:%-3d; status:%d; className:%-5s; methodName:%-5s\n", data.LineNo, data.Status, data.ClassName, data.MethodName)
}

func printVarData(values []*PdbgVarAttr) {
	for _, varData := range values {
		fmt.Printf("id:%-2d;name:%-6s;dataType:%-5s;size:%-3d;value:%s\n", varData.Id, varData.Name, varData.DataTypeName, varData.Size, strings.TrimSpace(varData.Value))
	}
}

package yasdb

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestDebugOutput(t *testing.T) {

	sql := `
	CREATE OR REPLACE PROCEDURE output() IS
	BEGIN
		DBMS_OUTPUT.PUT_LINE('1');
		DBMS_OUTPUT.PUT_LINE('2');
		DBMS_OUTPUT.PUT_LINE('3');
		DBMS_OUTPUT.PUT_LINE('4');
		DBMS_OUTPUT.PUT_LINE('5');
	END output;
	`

	callTemplate := `
	BEGIN
		"OUTPUT"();
	END;
	`

	proName := "OUTPUT"

	createProcedute(t, sql)
	oid, subid := queryObjIdAndSubId(t, proName)

	debug, err := NewPlsqlDebug(testDsn,
		WithDebugCallTempalate(callTemplate),
		WithDebugEnableOutput(),
		WithDebugOutputCacheSize(100000),
		WithDebugOutputMaxLineLen(2000),
	)
	defer func() {
		_ = debug.Abort()
		_ = debug.Close()
	}()
	if err != nil {
		t.Fatal(err)
	}
	if err := debug.Start(oid, subid); err != nil {
		t.Fatal(err)
	}
	outLines := []string{}
	expected := []string{"1", "2", "3", "4", "5"}
	for {
		if err := debug.StepNext(); err != nil {
			var yasErr *YasDBError
			// 调试结束
			if errors.As(err, &yasErr) && yasErr.Code == 8068 {
				break
			}
			t.Fatal(err)
		}
	}
	ouput, err := debug.GetOutput()
	if err != nil {
		t.Fatal(err)
	}
	outLines = append(outLines, ouput...)

	for i, line := range outLines {
		if line != expected[i] {
			t.Fatalf("expected line:\n %s \n actual line:\n %s \n,not equal", expected[i], line)
		}
	}

}

func TestPdbgGetOutput(t *testing.T) {
	sql := `
	CREATE OR REPLACE PROCEDURE output_capi() IS
	BEGIN
		DBMS_OUTPUT.PUT_LINE('hello');
		DBMS_OUTPUT.PUT_LINE('world');
		DBMS_OUTPUT.PUT_LINE('test');
	END output_capi;
	`

	callTemplate := `
	BEGIN
		"OUTPUT_CAPI"();
	END;
	`

	proName := "OUTPUT_CAPI"

	createProcedute(t, sql)
	oid, subid := queryObjIdAndSubId(t, proName)

	debug, err := NewPlsqlDebug(testDsn,
		WithDebugCallTempalate(callTemplate),
		WithDebugEnableOutput(),
		WithDebugOutputMaxLineLen(2000),
	)
	defer func() {
		_ = debug.Abort()
		_ = debug.Close()
	}()
	if err != nil {
		t.Fatal(err)
	}
	if err := debug.Start(oid, subid); err != nil {
		t.Fatal(err)
	}
	expected := []string{"hello\n", "world\n", "test\n"}
	lines := make([]string, 0, len(expected))
	i := 0
	for {
		if err := debug.StepNext(); err != nil {
			var yasErr *YasDBError
			if errors.As(err, &yasErr) && yasErr.Code == 8068 {
				break
			}
			t.Fatal(err)
		}
		fmt.Println("step next ", i)
		i++

		if _, err := debug.GetAllVarAttrs(); err != nil {
			fmt.Printf("GetAllVarAttrs failed: %v\n", err)

		}

		line, err := debug.PdbgGetOutput()
		if err != nil {
			if isLoadSymbolErr(err) {
				t.Skip("yasdb client does not support yacPdbgGetOutput:", err)
			}
			fmt.Printf("PdbgGetOutput failed: %v\n", err)
			continue
		}
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}

	if len(lines) != len(expected) {
		t.Fatalf("expected %d lines, got %d", len(expected), len(lines))
	}
	for i, line := range lines {
		if line != expected[i] {
			t.Fatalf("line %d: expected %q, got %q", i, expected[i], line)
		}
	}
}

func TestPdbgGetOutput66K(t *testing.T) {
	sql := `
	CREATE OR REPLACE PROCEDURE output_capi IS
		v_line VARCHAR2(1024);
	BEGIN
		-- 构造 1024 个字符的字符串
		v_line := RPAD('A', 1024, 'A');
		DBMS_OUTPUT.PUT_LINE('');

		FOR i IN 1 .. 66 LOOP
			DBMS_OUTPUT.PUT_LINE(v_line);
		END LOOP;
	END output_capi;
	`

	callTemplate := `
	BEGIN
		"OUTPUT_CAPI"();
	END;
	`

	proName := "OUTPUT_CAPI"

	createProcedute(t, sql)
	oid, subid := queryObjIdAndSubId(t, proName)

	debug, err := NewPlsqlDebug(testDsn,
		WithDebugCallTempalate(callTemplate),
		WithDebugEnableOutput(),
		WithDebugOutputMaxLineLen(66000),
	)
	defer func() {
		_ = debug.Abort()
		_ = debug.Close()
	}()
	if err != nil {
		t.Fatal(err)
	}
	if err := debug.Start(oid, subid); err != nil {
		t.Fatal(err)
	}
	i := 0
	for {
		if err := debug.StepNext(); err != nil {
			var yasErr *YasDBError
			if errors.As(err, &yasErr) && yasErr.Code == 8068 {
				break
			}
			t.Fatal(err)
		}
		fmt.Println("step next ", i)
		i++

		line, err := debug.PdbgGetOutput()
		if err != nil {
			if isLoadSymbolErr(err) {
				t.Skip("yasdb client does not support yacPdbgGetOutput:", err)
			}
			fmt.Printf("PdbgGetOutput failed: %v\n", err)
			continue
		}
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "A") && len(line) != (1024+1) {
			t.Fatalf("yacPdbgGetOutput seems no correct")
		}
	}
}

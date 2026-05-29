package yasdb

import (
	"errors"
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

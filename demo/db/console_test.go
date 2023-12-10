package db

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

type testConsoleTable struct {
	Id     int
	Field1 string
	Field2 string
	Field3 int
	Field4 float64
}

func TestConsole(t *testing.T) {
	mdb := &memoryDb{tables: sync.Map{}}

	table := NewTable("console-test").WithType(reflect.TypeOf(new(testConsoleTable)))
	err := mdb.CreateTable(table)
	if err != nil {
		t.Errorf("create table failed, err: %v", err)
	}

	tct := &testConsoleTable{Id: 1, Field1: "hello", Field2: "world", Field3: 10238, Field4: 12.343}
	err = mdb.Insert("console-test", 1, tct)
	if err != nil {
		t.Errorf("insert failed, err: %v", err)
	}

	result, err := mdb.ExecSql("select id,field1,field2,field3,field4 from console-test where id=1")
	if err != nil {
		t.Errorf("exec sql failed, err: %v", err)
	}

	render := NewTableRender(result)
	fmt.Println(render.Render())
}

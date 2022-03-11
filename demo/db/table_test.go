package db

import (
	"reflect"
	"testing"
)

type testRegion struct {
	Id   int
	Name string
}

func TestTable(t *testing.T) {
	table := NewTable("testRegion").WithType(reflect.TypeOf(new(testRegion)))
	table.Insert(2, &testRegion{Id: 2, Name: "beijing"})
	record := new(testRegion)
	if err := table.QueryByPrimaryKey(2, record); err != nil {
		t.Error(err)
	}
	if record.Name != "beijing" {
		t.Error("query failed, want beijing, got " + record.Name)
	}

	table.Update(2, &testRegion{Id: 2, Name: "shanghai"})
	if err := table.QueryByPrimaryKey(2, record); err != nil {
		t.Error(err)
	}
	if record.Name != "shanghai" {
		t.Error("query failed, want shanghai, got " + record.Name)
	}

	table.Delete(2)
	if err := table.QueryByPrimaryKey(2, record); err == nil {
		t.Error(err)
	}
}

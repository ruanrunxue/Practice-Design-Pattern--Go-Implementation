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
	table := NewTable("testRegion", NewRandomTableIteratorFactory())
	region := &testRegion{
		Id:   1,
		Name: "beijing",
	}
	if err := table.Insert(1, region); err != nil {
		t.Error(err)
	}
	table.Insert(2, &testRegion{Id: 2, Name: "beijing"})
	record := new(testRegion)
	if err := table.QueryByPrimaryKey(1, record); err != nil {
		t.Error(err)
	}
	records, err := table.QueryByConditions(reflect.TypeOf(new(testRegion)), NewCondition("name", "beijing"))
	if err != nil {
		t.Error(err)
	}
	if len(records) != 2 {
		t.Error("size invalid")
	}

	table.Update(2, &testRegion{Id: 2, Name: "shanghai"})
	records, err = table.QueryByConditions(reflect.TypeOf(new(testRegion)), NewCondition("name", "beijing"))
	if err != nil {
		t.Error(err)
	}
	if len(records) != 1 {
		t.Error("size invalid")
	}

	table.Delete(1)
	records, err = table.QueryByConditions(reflect.TypeOf(new(testRegion)), NewCondition("name", "beijing"))
	if err != nil {
		t.Error(err)
	}
	if len(records) != 0 {
		t.Error("size invalid")
	}
}

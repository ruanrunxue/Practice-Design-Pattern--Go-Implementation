package db

import (
	"reflect"
	"testing"
)

func TestDsl(t *testing.T) {
	db := NewMemoryDb()
	table := NewTable("region").
		WithType(reflect.TypeOf(new(testRegion))).
		WithTableIteratorFactory(NewRandomTableIteratorFactory())
	db.CreateTable(table)
	db.Insert("region", 1, &testRegion{Id: 1, Name: "beijing"})
	result, err := db.ExecDsl("select id,name from region where id=1")
	if err != nil {
		t.Error(err)
	}
	rs := result.ToMap()
	if rs["id"] != 1 {
		t.Errorf("id failed, want 1, got %v", rs["id"])
	}
	if rs["name"] != "beijing" {
		t.Errorf("name failed, want beijing, got %v", rs["name"])
	}
}

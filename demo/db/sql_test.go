package db

import (
	"reflect"
	"sync"
	"testing"
)

func TestSql(t *testing.T) {
	db := &memoryDb{tables: sync.Map{}}
	table := NewTable("region").
		WithType(reflect.TypeOf(new(testRegion))).
		WithTableIteratorFactory(NewRandomTableIteratorFactory())
	db.CreateTable(table)
	db.Insert("region", 1, &testRegion{Id: 1, Name: "beijing"})
	result, err := db.ExecSql("select id,name from region where id=1")
	if err != nil {
		t.Error(err)
	}
	rs := result.ToMap()
	if rs["id"] != 1 {
		t.Errorf("Id failed, want 1, got %v", rs["Id"])
	}
	if rs["name"] != "beijing" {
		t.Errorf("name failed, want beijing, got %v", rs["name"])
	}

	console := NewConsole(db)
	console.Start()
}

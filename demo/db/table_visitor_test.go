package db

import (
	"reflect"
	"testing"
)

type testRegionVisitor struct {
}

func (t *testRegionVisitor) Visit(table *Table) ([]interface{}, error) {
	iter := table.Iterator()
	var result []interface{}
	for iter.HasNext() {
		region := new(testRegion)
		if err := iter.Next(region); err != nil {
			return nil, err
		}
		if region.Name == "beijing" {
			result = append(result, region)
		}
	}
	return result, nil
}

func TestTableVisitor(t *testing.T) {
	table := NewTable("testRegion").WithType(reflect.TypeOf(new(testRegion)))
	table.Insert(1, &testRegion{Id: 1, Name: "beijing"})
	table.Insert(2, &testRegion{Id: 2, Name: "beijing"})
	table.Insert(3, &testRegion{Id: 3, Name: "guangdong"})
	result, err := table.Accept(&testRegionVisitor{})
	if err != nil {
		t.Error(err)
	}
	if len(result) != 2 {
		t.Errorf("visit failed, want 2, got %d", len(result))
	}
}

func TestFieldEqVisitor(t *testing.T) {
	table := NewTable("testRegion").WithType(reflect.TypeOf(new(testRegion)))
	table.Insert(1, &testRegion{Id: 1, Name: "beijing"})
	table.Insert(2, &testRegion{Id: 2, Name: "beijing"})
	table.Insert(3, &testRegion{Id: 3, Name: "guangdong"})

	visitor1 := &FieldEqVisitor{field: "name", value: "beijing"}
	result, err := table.Accept(visitor1)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 2 {
		t.Errorf("visit failed, want 2, got %d", len(result))
	}

	visitor2 := &FieldEqVisitor{field: "id", value: 4}
	result, err = table.Accept(visitor2)
	if err != ErrRecordNotFound {
		t.Errorf("visit failed, want ErrRecordNotFound, got %v", err)
	}
}

func TestFieldEqVisitorFunc(t *testing.T) {
	table := NewTable("testRegion").WithType(reflect.TypeOf(new(testRegion)))
	table.Insert(1, &testRegion{Id: 1, Name: "beijing"})
	table.Insert(2, &testRegion{Id: 2, Name: "beijing"})
	table.Insert(3, &testRegion{Id: 3, Name: "guangdong"})

	result, err := table.AcceptFunc(NewFieldEqVisitorFunc("name", "beijing"))
	if err != nil {
		t.Error(err)
	}
	if len(result) != 2 {
		t.Errorf("visit failed, want 2, got %d", len(result))
	}

	result, err = table.AcceptFunc(NewFieldEqVisitorFunc("id", 4))
	if err != ErrRecordNotFound {
		t.Errorf("visit failed, want ErrRecordNotFound, got %v", err)
	}
}

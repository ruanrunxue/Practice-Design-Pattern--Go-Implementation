package model

import (
	"demo/db"
	"reflect"
	"testing"
)

func TestRegion(t *testing.T) {
	mdb := db.MemoryDbInstance()
	defer db.MemoryDbInstance().Clear()

	table := db.NewTable("region-test").WithType(reflect.TypeOf(new(Region)))
	mdb.CreateTable(table)

	region := NewRegion("1")
	region.Name = "region-1"
	region.Country = "CN"
	mdb.Insert("region-test", "1", region)

	result := new(Region)
	mdb.Query("region-test", "1", result)

	if !reflect.DeepEqual(result, region) {
		t.Errorf("want %+v, got %+v", region, result)
	}
}

package model

import (
	"demo/db"
	"reflect"
	"testing"
)

func TestServiceProfile(t *testing.T) {
	mdb := db.MemoryDbInstance()
	defer db.MemoryDbInstance().Clear()

	table := db.NewTable("profile-test").WithType(reflect.TypeOf(new(ServiceProfile)))
	mdb.CreateTable(table)

	profile := NewServiceProfileBuilder().
		WithId("svc1").WithType("svc").WithStatus(Normal).
		WithEndpoint("192.168.0.1", 80).WithRegionId("1").
		WithPriority(1).WithLoad(100).Build()
	err := mdb.Insert("profile-test", "svc1", profile)
	if err != nil {
		t.Fatal(err)
	}

	visitor := NewServiceProfileVisitor("svc1", "")
	result, _ := mdb.QueryByVisitor("profile-test", visitor)
	if !reflect.DeepEqual(result[0], profile) {
		t.Errorf("want %+v, got %+v", profile, result)
	}
}

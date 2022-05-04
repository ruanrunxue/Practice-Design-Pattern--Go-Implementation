package model

import (
	"demo/db"
	"reflect"
	"testing"
)

func TestServiceProfile(t *testing.T) {
	mdb := db.MemoryDbInstance()
	defer db.MemoryDbInstance().Clear()

	table := db.NewTable("profile-test").WithType(reflect.TypeOf(new(ServiceProfileRecord)))
	mdb.CreateTable(table)

	profile := NewServiceProfileBuilder().
		WithId("svc1").WithType("svc").WithStatus(Normal).
		WithEndpoint("192.168.0.1", 80).WithRegion(NewRegion("1")).
		WithPriority(1).WithLoad(100).Build()
	err := mdb.Insert("profile-test", "svc1", profile.ToTableRecord())
	if err != nil {
		t.Fatal(err)
	}

	visitor := NewServiceProfileVisitor("svc1", "")
	result, _ := mdb.QueryByVisitor("profile-test", visitor)
	got := result[0].(*ServiceProfileRecord).ToServiceProfile()
	if !reflect.DeepEqual(got, profile) {
		t.Errorf("want %+v, got %+v", got, profile)
	}
}

func TestNewServiceProfile(t *testing.T) {
	profile := NewServiceProfile("svc1", "svc",
		Status(Normal),
		Endpoint("192.168.10.10", 80),
		SvcRegion("region2", "guangdong", "China"),
		Priority(2),
		Load(300))
	if profile.Priority != 2 {
		t.Errorf("want 2, got %d", profile.Priority)
	}
}

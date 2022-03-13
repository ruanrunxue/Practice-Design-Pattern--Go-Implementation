package model

import (
	"demo/db"
	"reflect"
	"testing"
)

func TestSubscription(t *testing.T) {
	mdb := db.MemoryDbInstance()
	defer db.MemoryDbInstance().Clear()

	table := db.NewTable("sub-test").WithType(reflect.TypeOf(new(Subscription)))
	mdb.CreateTable(table)

	sub1 := NewSubscription("1")
	sub1.SrcSvcId = "svc1"
	sub1.TargetSvcId = "svc2"
	sub1.TargetSvcType = "svc"
	mdb.Insert("sub-test", "1", sub1)

	sub2 := NewSubscription("2")
	sub2.SrcSvcId = "svc2"
	sub2.TargetSvcId = "svc3"
	sub2.TargetSvcType = "svc"
	mdb.Insert("sub-test", "2", sub2)

	visitor := NewSubscriptionVisitor("", "svc")
	result, _ := mdb.QueryByVisitor("sub-test", visitor)
	if len(result) != 2 {
		t.Errorf("want 2 got %d", len(result))
	}
}

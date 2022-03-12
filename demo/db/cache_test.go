package db

import (
	"reflect"
	"sync"
	"testing"
)

func TestCacheProxy(t *testing.T) {
	cache := NewCacheProxy(&memoryDb{tables: sync.Map{}})
	table := NewTable("region").
		WithType(reflect.TypeOf(new(testRegion))).
		WithTableIteratorFactory(NewRandomTableIteratorFactory())
	cache.CreateTable(table)
	table.Insert(1, &testRegion{Id: 1, Name: "region"})

	result := new(testRegion)
	cache.Query("region", 1, result)
	if cache.Miss() != 1 {
		t.Errorf("cache miss error, want 1 got %d\n", cache.Miss())
	}
	cache.Query("region", 1, result)
	if cache.Hit() != 1 {
		t.Errorf("cache hit error, want 1 got %d\n", cache.Hit())
	}
}

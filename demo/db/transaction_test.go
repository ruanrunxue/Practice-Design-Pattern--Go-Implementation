package db

import (
	"reflect"
	"sync"
	"testing"
)

func TestTransaction_Success(t *testing.T) {
	db := &memoryDb{tables: sync.Map{}}
	db.CreateTable(NewTable("region1").
		WithType(reflect.TypeOf(new(testRegion))).
		WithTableIteratorFactory(NewRandomTableIteratorFactory()))
	db.CreateTable(NewTable("region2").
		WithType(reflect.TypeOf(new(testRegion))).
		WithTableIteratorFactory(NewRandomTableIteratorFactory()))
	transaction := db.CreateTransaction("region_trans")
	transaction.Begin()
	err := transaction.Exec(NewInsertCmd("region1").
		WithPrimaryKey(1).
		WithRecord(&testRegion{Id: 1, Name: "beijing"}))
	if err != nil {
		t.Error(err)
	}
	err = transaction.Exec(NewInsertCmd("region2").
		WithPrimaryKey(2).
		WithRecord(&testRegion{Id: 2, Name: "shanghai"}))
	if err != nil {
		t.Error(err)
	}
	err = transaction.Commit()
	if err != nil {
		t.Error(err)
	}

	result := new(testRegion)
	db.Query("region1", 1, result)
	if result.Name != "beijing" {
		t.Error(result.Name)
	}
}

func TestTransaction_Failed(t *testing.T) {
	db := &memoryDb{tables: sync.Map{}}
	db.CreateTable(NewTable("region1").
		WithType(reflect.TypeOf(new(testRegion))).
		WithTableIteratorFactory(NewRandomTableIteratorFactory()))
	transaction := db.CreateTransaction("region_trans")
	transaction.Begin()
	err := transaction.Exec(NewInsertCmd("region1").
		WithPrimaryKey(1).
		WithRecord(&testRegion{Id: 1, Name: "beijing"}))
	if err != nil {
		t.Error(err)
	}
	err = transaction.Exec(NewInsertCmd("region2").
		WithPrimaryKey(2).
		WithRecord(&testRegion{Id: 2, Name: "shanghai"}))
	if err != nil {
		t.Error(err)
	}
	err = transaction.Commit()
	if err == nil {
		t.Error("commit failed")
	}

	result := new(testRegion)
	err = db.Query("region1", 1, result)
	if err == nil {
		t.Error("transaction failed")
	}

}

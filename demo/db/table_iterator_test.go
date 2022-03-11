package db

import "testing"

func TestRandomTableIterator(t *testing.T) {
	table := NewTable("testRegion", NewRandomTableIteratorFactory())
	table.Insert(1, &testRegion{Id: 1, Name: "beijing"})
	table.Insert(2, &testRegion{Id: 2, Name: "shanghai"})
	table.Insert(3, &testRegion{Id: 3, Name: "guangdong"})
	iter := table.Iterator()
	for i := 0; i < 3; i++ {
		if !iter.HasNext() {
			t.Error("records size error")
		}
		region := new(testRegion)
		if err := iter.Next(region); err != nil {
			t.Error(err)
		}
	}
}

func TestSortTableIterator(t *testing.T) {
	table := NewTable("testRegion", NewRandomTableIteratorFactory())
	table.Insert(3, &testRegion{Id: 3, Name: "beijing"})
	table.Insert(1, &testRegion{Id: 1, Name: "shanghai"})
	table.Insert(2, &testRegion{Id: 2, Name: "guangdong"})
	iter := table.Iterator()
	region1 := new(testRegion)
	iter.Next(region1)
	if region1.Id != 1 {
		t.Error("region1 sort failed")
	}
	region2 := new(testRegion)
	iter.Next(region2)
	if region2.Id != 2 {
		t.Error("region2 sort failed")
	}
	region3 := new(testRegion)
	iter.Next(region3)
	if region3.Id != 3 {
		t.Error("region3 sort failed")
	}
}

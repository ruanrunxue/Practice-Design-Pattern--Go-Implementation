package db

import (
	"math/rand"
	"sort"
	"time"
)

/*
迭代器模式
*/

// TableIterator 表迭代器接口
type TableIterator interface {
	HasNext() bool
	Next(result interface{}) error
}

// TableIteratorFactory 表迭代器工厂
type TableIteratorFactory interface {
	Create(table *Table) TableIterator
}

// tableIteratorImpl 迭代器基类
type tableIteratorImpl struct {
	records []record
	cursor  int
}

func (r *tableIteratorImpl) HasNext() bool {
	return r.cursor < len(r.records)
}

func (r *tableIteratorImpl) Next(next interface{}) error {
	record := r.records[r.cursor]
	r.cursor++
	if err := record.convertByValue(next); err != nil {
		return err
	}
	return nil
}

type randomTableIteratorFactory struct{}

func (r *randomTableIteratorFactory) Create(table *Table) TableIterator {
	var records []record
	for _, r := range table.records {
		records = append(records, r)
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(records), func(i, j int) {
		records[i], records[j] = records[j], records[i]
	})
	return &tableIteratorImpl{
		records: records,
		cursor:  0,
	}
}

func NewRandomTableIteratorFactory() *randomTableIteratorFactory {
	return &randomTableIteratorFactory{}
}

// Comparator 如果i<j返回true，否则返回false
type Comparator func(i, j interface{}) bool

// records 辅助record记录根据主键排序
type records struct {
	comparator Comparator
	rs         []record
}

func newRecords(rs []record, comparator Comparator) *records {
	return &records{
		comparator: comparator,
		rs:         rs,
	}
}

func (r *records) Len() int {
	return len(r.rs)
}

func (r *records) Less(i, j int) bool {
	return r.comparator(r.rs[i].primaryKey, r.rs[j].primaryKey)
}

func (r *records) Swap(i, j int) {
	tmp := r.rs[i]
	r.rs[i] = r.rs[j]
	r.rs[j] = tmp
}

// sortedTableIteratorFactory 根据主键进行排序，排序逻辑由Comparator定义
type sortedTableIteratorFactory struct {
	comparator Comparator
}

func (s *sortedTableIteratorFactory) Create(table *Table) TableIterator {
	var records []record
	for _, r := range table.records {
		records = append(records, r)
	}
	sort.Sort(newRecords(records, s.comparator))
	return &tableIteratorImpl{
		records: records,
		cursor:  0,
	}
}

func NewSortedTableIteratorFactory(comparator Comparator) *sortedTableIteratorFactory {
	return &sortedTableIteratorFactory{comparator: comparator}
}

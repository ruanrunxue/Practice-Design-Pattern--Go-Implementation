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
	Next(next interface{}) error
}

// TableIteratorFactory 表迭代器工厂
type TableIteratorFactory interface {
	Create(table *Table) TableIterator
}

// tableIteratorImpl 随机迭代器
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

// Less 如果i<j返回ture，否则返回false
type Less func(i, j interface{}) bool

// records 辅助record记录根据主键排序
type records struct {
	less Less
	rs   []record
}

func newRecords(rs []record, less Less) *records {
	return &records{
		less: less,
		rs:   rs,
	}
}

func (r *records) Len() int {
	return len(r.rs)
}

func (r *records) Less(i, j int) bool {
	return r.less(r.rs[i].primaryKey, r.rs[j].primaryKey)
}

func (r *records) Swap(i, j int) {
	tmp := r.rs[i]
	r.rs[i] = r.rs[j]
	r.rs[j] = tmp
}

// sortedTableIteratorFactory 根据主键进行排序，排序逻辑由Comparable定义
type sortedTableIteratorFactory struct {
	less Less
}

func (s *sortedTableIteratorFactory) Create(table *Table) TableIterator {
	var records []record
	for _, r := range table.records {
		records = append(records, r)
	}
	sort.Sort(newRecords(records, s.less))
	return &tableIteratorImpl{
		records: records,
		cursor:  0,
	}
}

func NewSortedTableIteratorFactory(less Less) *sortedTableIteratorFactory {
	return &sortedTableIteratorFactory{less: less}
}

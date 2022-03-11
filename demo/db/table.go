package db

import (
	"reflect"
	"strings"
)

// Table 数据表定义
type Table struct {
	name            string
	recordType      reflect.Type
	records         map[interface{}]record
	iteratorFactory TableIteratorFactory // 默认使用随机迭代器
}

func NewTable(name string) *Table {
	return &Table{
		name:            name,
		records:         make(map[interface{}]record),
		iteratorFactory: NewRandomTableIteratorFactory(),
	}
}

func (t *Table) WithType(recordType reflect.Type) *Table {
	t.recordType = recordType
	return t
}

func (t *Table) WithTableIteratorFactory(iteratorFactory TableIteratorFactory) *Table {
	t.iteratorFactory = iteratorFactory
	return t
}

func (t *Table) Name() string {
	return strings.ToLower(t.name)
}

func (t *Table) QueryByPrimaryKey(key interface{}, value interface{}) error {
	record, ok := t.records[key]
	if !ok {
		return ErrRecordNotFound
	}
	return record.convertByValue(value)
}

func (t *Table) Insert(key interface{}, value interface{}) error {
	if _, ok := t.records[key]; ok {
		return ErrPrimaryKeyConflict
	}
	record, err := recordFrom(key, value)
	if err != nil {
		return err
	}
	t.records[key] = record
	return nil
}

func (t *Table) Update(key interface{}, value interface{}) error {
	if _, ok := t.records[key]; !ok {
		return ErrRecordNotFound
	}
	record, err := recordFrom(key, value)
	if err != nil {
		return err
	}
	t.records[key] = record
	return nil
}

func (t *Table) Delete(key interface{}) error {
	if _, ok := t.records[key]; !ok {
		return ErrRecordNotFound
	}
	delete(t.records, key)
	return nil
}

func (t *Table) Iterator() TableIterator {
	return t.iteratorFactory.Create(t)
}

func (t *Table) Accept(visitor TableVisitor) ([]interface{}, error) {
	return visitor.Visit(t)
}

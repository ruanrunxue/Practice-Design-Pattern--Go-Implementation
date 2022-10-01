package db

import "reflect"

/*
访问者模式
*/

type TableVisitor interface {
	Visit(table *Table) ([]interface{}, error)
}

// FieldEqVisitor 根据给定的字段名和值，找到相应的记录
type FieldEqVisitor struct {
	field string
	value interface{}
}

func (f *FieldEqVisitor) Visit(table *Table) ([]interface{}, error) {
	result := make([]interface{}, 0)
	idx, ok := table.metadata[f.field]
	if !ok {
		return nil, ErrRecordNotFound
	}
	for _, r := range table.records {
		if reflect.DeepEqual(r.values[idx], f.value) {
			result = append(result, r)
		}
	}
	if len(result) == 0 {
		return nil, ErrRecordNotFound
	}
	return result, nil
}

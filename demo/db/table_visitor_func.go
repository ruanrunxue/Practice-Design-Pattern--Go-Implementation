package db

import "reflect"

func (t *Table) AcceptFunc(visitor TableVisitorFunc) ([]interface{}, error) {
	return visitor(t)
}

type TableVisitorFunc func(table *Table) ([]interface{}, error)

func NewFieldEqVisitorFunc(field string, value interface{}) TableVisitorFunc {
	return func(table *Table) ([]interface{}, error) {
		result := make([]interface{}, 0)
		idx, ok := table.metadata[field]
		if !ok {
			return nil, ErrRecordNotFound
		}
		for _, r := range table.records {
			if reflect.DeepEqual(r.values[idx], value) {
				result = append(result, r)
			}
		}
		if len(result) == 0 {
			return nil, ErrRecordNotFound
		}
		return result, nil
	}
}

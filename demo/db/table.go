package db

import "reflect"

// Condition 查询条件
type Condition struct {
	field string
	value interface{}
}

func NewCondition(field string, value interface{}) Condition {
	return Condition{field: field, value: value}
}

// Index 索引，第一个key为field name，第二个key为value，value为对应的主键
type Index map[string]map[interface{}][]interface{}

func (i *Index) add(r record) {
	for field, cursor := range r.fields {
		idx, ok := (*i)[field]
		if !ok {
			continue
		}
		idx[r.values[cursor]] = append(idx[r.values[cursor]], r.primaryKey)
	}
}

func (i *Index) get(field string, value interface{}) []interface{} {
	idxs, ok := (*i)[field]
	if !ok {
		return []interface{}{}
	}
	cursors, ok := idxs[value]
	if !ok {
		return []interface{}{}
	}
	return cursors
}

// Table 数据表定义
type Table struct {
	name    string
	index   Index
	records map[interface{}]record
}

func NewTable(name string) *Table {
	return &Table{
		name:    name,
		index:   make(map[string]map[interface{}][]interface{}),
		records: make(map[interface{}]record),
	}
}

// CreateIndex 为字段field创建索引
func (t *Table) CreateIndex(field string) {
	t.index[field] = make(map[interface{}][]interface{})
}

func (t *Table) QueryByPrimaryKey(key interface{}, value interface{}) error {
	record, ok := t.records[key]
	if !ok {
		return ErrRecordNotFound
	}
	return record.convertByValue(value)
}

func (t *Table) QueryByConditions(valType reflect.Type, condition Condition) ([]interface{}, error) {
	keys := t.index.get(condition.field, condition.value)
	iVals, err := t.queryByKeys(valType, keys)
	if err != nil {
		return []interface{}{}, err
	}
	if len(iVals) != 0 {
		return iVals, nil
	}
	var result []interface{}
	for _, v := range t.records {
		cursor, ok := v.fields[condition.field]
		if !ok {
			return []interface{}{}, ErrRecordNotFound
		}
		if reflect.DeepEqual(condition.value, v.values[cursor]) {
			r, err := v.convertByType(valType)
			if err != nil {
				return []interface{}{}, err
			}
			result = append(result, r)
		}
	}
	return result, nil
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
	// 更新索引
	t.index.add(record)
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
	// 更新索引
	t.index.add(record)
	return nil
}

func (t *Table) Delete(key interface{}) error {
	if _, ok := t.records[key]; !ok {
		return ErrRecordNotFound
	}
	delete(t.records, key)
	return nil
}

func (t *Table) queryByKeys(valType reflect.Type, keys []interface{}) ([]interface{}, error) {
	var result []interface{}
	for _, k := range keys {
		r, err := t.records[k].convertByType(valType)
		if err != nil {
			return []interface{}{}, err
		}
		result = append(result, r)
	}
	return result, nil
}

package db

import (
	"reflect"
)

type record struct {
	primaryKey interface{}
	values     []interface{} // 存储属性值
}

func recordFrom(key interface{}, value interface{}) (r record, e error) {
	defer func() {
		if err := recover(); err != nil {
			r = record{}
			e = ErrRecordTypeInvalid
		}
	}()
	vVal := reflect.ValueOf(value)
	if vVal.Type().Kind() == reflect.Pointer {
		vVal = vVal.Elem()
	}
	record := record{
		primaryKey: key,
		values:     make([]interface{}, vVal.NumField()),
	}
	for i := 0; i < vVal.NumField(); i++ {
		fieldVal := vVal.Field(i)
		record.values[i] = fieldVal.Interface()
	}
	return record, nil
}

func (r record) convertByValue(result interface{}) (e error) {
	defer func() {
		if err := recover(); err != nil {
			e = ErrRecordTypeInvalid
		}
	}()
	rType := reflect.TypeOf(result)
	rVal := reflect.ValueOf(result)
	if rType.Kind() == reflect.Pointer {
		rType = rType.Elem()
		rVal = rVal.Elem()
	}
	for i := 0; i < rType.NumField(); i++ {
		field := rVal.Field(i)
		field.Set(reflect.ValueOf(r.values[i]))
	}
	return nil
}

func (r record) convertByType(rType reflect.Type) (result interface{}, e error) {
	defer func() {
		if err := recover(); err != nil {
			e = ErrRecordTypeInvalid
		}
	}()
	if rType.Kind() == reflect.Pointer {
		rType = rType.Elem()
	}
	rVal := reflect.New(rType)
	return rVal, nil
}

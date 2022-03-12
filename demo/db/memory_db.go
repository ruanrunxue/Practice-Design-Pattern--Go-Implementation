package db

import (
	"strings"
	"sync"
)

var memoryDbInstance = &memoryDb{tables: sync.Map{}}

// memoryDb 内存数据库
type memoryDb struct {
	tables sync.Map // key为tableName，value为table
}

func MemoryDbInstance() *memoryDb {
	return memoryDbInstance
}

func (m *memoryDb) CreateTable(t *Table) error {
	if _, ok := m.tables.Load(t.Name()); ok {
		return ErrTableAlreadyExist
	}
	m.tables.Store(t.Name(), t)
	return nil
}

func (m *memoryDb) CreateTableIfNotExist(t *Table) error {
	if _, ok := m.tables.Load(t.Name()); ok {
		return nil
	}
	m.tables.Store(t.Name(), t)
	return nil
}

func (m *memoryDb) DeleteTable(tableName string) error {
	if _, ok := m.tables.Load(tableName); !ok {
		return ErrTableNotExist
	}
	m.tables.Delete(tableName)
	return nil
}

func (m *memoryDb) Query(tableName string, primaryKey interface{}, result interface{}) error {
	table, ok := m.tables.Load(tableName)
	if !ok {
		return ErrTableNotExist
	}
	return table.(*Table).QueryByPrimaryKey(primaryKey, result)
}

func (m *memoryDb) QueryByVisitor(tableName string, visitor TableVisitor) ([]interface{}, error) {
	table, ok := m.tables.Load(tableName)
	if !ok {
		return nil, ErrTableNotExist
	}
	return table.(*Table).Accept(visitor)
}

func (m *memoryDb) Insert(tableName string, primaryKey interface{}, record interface{}) error {
	table, ok := m.tables.Load(tableName)
	if !ok {
		return ErrTableNotExist
	}
	return table.(*Table).Insert(primaryKey, record)
}

func (m *memoryDb) Update(tableName string, primaryKey interface{}, record interface{}) error {
	table, ok := m.tables.Load(tableName)
	if !ok {
		return ErrTableNotExist
	}
	return table.(*Table).Update(primaryKey, record)
}

func (m *memoryDb) Delete(tableName string, primaryKey interface{}) error {
	table, ok := m.tables.Load(tableName)
	if !ok {
		return ErrTableNotExist
	}
	return table.(*Table).Delete(primaryKey)
}

func (m *memoryDb) CreateTransaction(name string) *Transaction {
	return NewTransaction(name, m)
}

func (m *memoryDb) ExecDsl(dsl string) (*DslResult, error) {
	ctx := NewDslContext()
	express := &CompoundExpression{dsl: dsl}
	if err := express.Interpret(ctx); err != nil {
		return nil, ErrDslInvalidGrammar
	}
	table, ok := m.tables.Load(ctx.TableName())
	if !ok {
		return nil, ErrTableNotExist
	}
	record, ok := table.(*Table).records[ctx.PrimaryKey()]
	if !ok {
		return nil, ErrRecordNotFound
	}
	result := NewDslResult()
	for _, f := range ctx.Fields() {
		field := strings.ToLower(f)
		if idx, ok := record.fields[field]; ok {
			result.Add(field, record.values[idx])
		}
	}
	return result, nil
}

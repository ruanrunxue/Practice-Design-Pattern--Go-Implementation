package db

import (
	"strings"
	"sync"
)

// MemoryDb 内存数据库
type MemoryDb struct {
	tables sync.Map // key为tableName，value为table
}

func NewMemoryDb() *MemoryDb {
	return &MemoryDb{tables: sync.Map{}}
}

func (m *MemoryDb) CreateTable(t *Table) error {
	if _, ok := m.tables.Load(t.Name()); ok {
		return ErrTableAlreadyExist
	}
	m.tables.Store(t.Name(), t)
	return nil
}

func (m *MemoryDb) CreateTableIfNotExist(t *Table) error {
	if _, ok := m.tables.Load(t.Name()); ok {
		return nil
	}
	m.tables.Store(t.Name(), t)
	return nil
}

func (m *MemoryDb) DeleteTable(tableName string) error {
	if _, ok := m.tables.Load(tableName); !ok {
		return ErrTableNotExist
	}
	m.tables.Delete(tableName)
	return nil
}

func (m *MemoryDb) Query(tableName string, primaryKey interface{}, result interface{}) error {
	table, ok := m.tables.Load(tableName)
	if !ok {
		return ErrTableNotExist
	}
	return table.(*Table).QueryByPrimaryKey(primaryKey, result)
}

func (m *MemoryDb) QueryByVisitor(tableName string, visitor TableVisitor) ([]interface{}, error) {
	table, ok := m.tables.Load(tableName)
	if !ok {
		return nil, ErrTableNotExist
	}
	return table.(*Table).Accept(visitor)
}

func (m *MemoryDb) Insert(tableName string, primaryKey interface{}, record interface{}) error {
	table, ok := m.tables.Load(tableName)
	if !ok {
		return ErrTableNotExist
	}
	return table.(*Table).Insert(primaryKey, record)
}

func (m *MemoryDb) Update(tableName string, primaryKey interface{}, record interface{}) error {
	table, ok := m.tables.Load(tableName)
	if !ok {
		return ErrTableNotExist
	}
	return table.(*Table).Update(primaryKey, record)
}

func (m *MemoryDb) Delete(tableName string, primaryKey interface{}) error {
	table, ok := m.tables.Load(tableName)
	if !ok {
		return ErrTableNotExist
	}
	return table.(*Table).Delete(primaryKey)
}

func (m *MemoryDb) CreateTransaction(name string) *Transaction {
	return NewTransaction(name, m)
}

func (m *MemoryDb) ExecDsl(dsl string) (*DslResult, error) {
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

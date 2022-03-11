package db

// Db 数据库抽象接口
type Db interface {
	CreateTable(t *Table) error
	CreateTableIfNotExist(t *Table) error
	DeleteTable(tableName string) error

	Query(tableName string, primaryKey interface{}, result interface{}) error
	QueryByVisitor(tableName string, visitor TableVisitor) ([]interface{}, error)
	Insert(tableName string, primaryKey interface{}, record interface{}) error
	Update(tableName string, primaryKey interface{}, record interface{}) error
	Delete(tableName string, primaryKey interface{}) error
}

package db

/**
 * 依赖倒置原则（DIP）：
 * 1、高层模块不应该依赖低层模块，两者都应该依赖抽象
 * 2、抽象不应该依赖细节，细节应该依赖抽象
 * DIP并不是说高层模块是只能依赖抽象接口，它的本意应该是依赖稳定的接口/抽象类/具象类。
 * 如果一个具象类是稳定的，比如Java中的String，那么高层模块依赖它也没有问题；
 * 相反，如果一个抽象接口是不稳定的，经常变化，那么高层模块依赖该接口也是违反DIP的，这时候应该思考下接口是否抽象合理。
 * 例子：
 * 抽象出Db接口，避免高层应用依赖具体的Db实现（比如MemoryDb），符合DIP
 */

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

	CreateTransaction(name string) *Transaction

	ExecDsl(dsl string) (*DslResult, error)
}

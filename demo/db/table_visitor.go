package db

/*
访问者模式
*/

type TableVisitor interface {
	Visit(table *Table) ([]interface{}, error)
}

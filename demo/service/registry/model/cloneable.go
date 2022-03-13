package model

/*
原型模式
*/

// Cloneable 原型复制接口
type Cloneable interface {
	Clone() Cloneable
}

package db

import (
	"strconv"
	"strings"
)

/*
解释器模式
*/

// DslContext DSL解析器上下文，保存各个表达式解析的中间结果
// 当前只支持基于主键的查询DSL语句
type DslContext struct {
	tableName  string
	fields     []string
	primaryKey interface{}
}

func NewDslContext() *DslContext {
	return &DslContext{}
}

func (d *DslContext) TableName() string {
	return d.tableName
}

func (d *DslContext) SetTableName(tableName string) {
	d.tableName = tableName
}

func (d *DslContext) Fields() []string {
	return d.fields
}

func (d *DslContext) SetFields(fields []string) {
	d.fields = fields
}

func (d *DslContext) PrimaryKey() interface{} {
	return d.primaryKey
}

func (d *DslContext) SetPrimaryKey(primaryKey interface{}) {
	d.primaryKey = primaryKey
}

// DslExpression DSL表达式抽象接口，每个词、符号和句子都属于表达式
type DslExpression interface {
	Interpret(ctx *DslContext) error
}

// SelectExpression select语句解析逻辑，select关键字后面跟的为field，以,分割，比如select id,name
type SelectExpression struct {
	fields string
}

func (s *SelectExpression) Interpret(ctx *DslContext) error {
	fields := strings.Split(s.fields, ",")
	if len(fields) == 0 {
		return ErrDslInvalidGrammar
	}
	ctx.SetFields(fields)
	return nil
}

// FromExpression from语句解析逻辑，from关键字后面跟的为表名，比如from regionTable1
type FromExpression struct {
	tableName string
}

func (f *FromExpression) Interpret(ctx *DslContext) error {
	if f.tableName == "" {
		return ErrDslInvalidGrammar
	}
	ctx.SetTableName(f.tableName)
	return nil
}

type WhereExpression struct {
	condition string
}

func (w *WhereExpression) Interpret(ctx *DslContext) error {
	vals := strings.Split(w.condition, "=")
	if len(vals) != 2 {
		return ErrDslInvalidGrammar
	}
	if strings.Contains(vals[1], "'") {
		ctx.SetPrimaryKey(strings.Trim(vals[1], "'"))
		return nil
	}
	if val, err := strconv.Atoi(vals[1]); err == nil {
		ctx.SetPrimaryKey(val)
		return nil
	}
	return ErrDslInvalidGrammar
}

// CompoundExpression DSL语句解释器，DSL固定为select xxx,xxx,xxx from xxx where xxx=xxx 的固定格式
// 例子：select regionId from regionTable where regionId=1
type CompoundExpression struct {
	dsl string
}

func (c *CompoundExpression) Interpret(ctx *DslContext) error {
	childs := strings.Split(c.dsl, " ")
	if len(childs) != 6 {
		return ErrDslInvalidGrammar
	}
	for i := 0; i < len(childs); i++ {
		switch strings.ToLower(childs[i]) {
		case "select":
			i++
			express := &SelectExpression{fields: childs[i]}
			if err := express.Interpret(ctx); err != nil {
				return err
			}
		case "from":
			i++
			express := &FromExpression{tableName: childs[i]}
			if err := express.Interpret(ctx); err != nil {
				return err
			}
		case "where":
			i++
			express := &WhereExpression{condition: childs[i]}
			if err := express.Interpret(ctx); err != nil {
				return err
			}
		default:
			return ErrDslInvalidGrammar
		}
	}
	return nil
}

// DslResult Dsl语句执行返回的结果
type DslResult struct {
	results map[string]interface{}
}

func NewDslResult() *DslResult {
	return &DslResult{results: make(map[string]interface{})}
}

func (d *DslResult) Add(field string, record interface{}) {
	d.results[field] = record
}

func (d *DslResult) ToMap() map[string]interface{} {
	return d.results
}

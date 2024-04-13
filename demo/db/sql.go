package db

import (
    "strconv"
    "strings"
)

/*
解释器模式
*/

// SqlContext SQL解析器上下文，保存各个表达式解析的中间结果
// 当前只支持基于主键的查询SQL语句
type SqlContext struct {
    tableName  string
    fields     []string
    primaryKey interface{}
}

func NewSqlContext() *SqlContext {
    return &SqlContext{}
}

func (s *SqlContext) TableName() string {
    return s.tableName
}

func (s *SqlContext) SetTableName(tableName string) {
    s.tableName = tableName
}

func (s *SqlContext) Fields() []string {
    return s.fields
}

func (s *SqlContext) SetFields(fields []string) {
    s.fields = fields
}

func (s *SqlContext) PrimaryKey() interface{} {
    return s.primaryKey
}

func (s *SqlContext) SetPrimaryKey(primaryKey interface{}) {
    s.primaryKey = primaryKey
}

// SqlExpression Sql表达式抽象接口，每个词、符号和句子都属于表达式
type SqlExpression interface {
    Interpret(ctx *SqlContext) error
}

// SelectExpression select语句解析逻辑，select关键字后面跟的为field，以,分割，比如select Id,name
type SelectExpression struct {
    fields string
}

func (s *SelectExpression) Interpret(ctx *SqlContext) error {
    fields := strings.Split(s.fields, ",")
    if len(fields) == 0 {
        return ErrSqlInvalidGrammar
    }
    ctx.SetFields(fields)
    return nil
}

// FromExpression from语句解析逻辑，from关键字后面跟的为表名，比如from regionTable1
type FromExpression struct {
    tableName string
}

func (f *FromExpression) Interpret(ctx *SqlContext) error {
    if f.tableName == "" {
        return ErrSqlInvalidGrammar
    }
    ctx.SetTableName(f.tableName)
    return nil
}

type WhereExpression struct {
    condition string
}

func (w *WhereExpression) Interpret(ctx *SqlContext) error {
    vals := strings.Split(w.condition, "=")
    if len(vals) != 2 {
        return ErrSqlInvalidGrammar
    }
    if strings.Contains(vals[1], "'") {
        ctx.SetPrimaryKey(strings.Trim(vals[1], "'"))
        return nil
    }
    if val, err := strconv.Atoi(vals[1]); err == nil {
        ctx.SetPrimaryKey(val)
        return nil
    }
    return ErrSqlInvalidGrammar
}

// CompoundExpression SQL语句解释器，SQL固定为select xxx,xxx,xxx from xxx where xxx=xxx; 的固定格式
// 例子：select regionId from regionTable where regionId=1
type CompoundExpression struct {
    sql string
}

func (c *CompoundExpression) Interpret(ctx *SqlContext) error {
    childs := strings.Split(c.sql, " ")
    if len(childs) != 6 {
        return ErrSqlInvalidGrammar
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
            return ErrSqlInvalidGrammar
        }
    }
    return nil
}

// SqlResult SQL语句执行返回的结果
type SqlResult struct {
    fields []string
    vals   []interface{}
}

func NewSqlResult() *SqlResult {
    return &SqlResult{
        fields: make([]string, 0),
        vals:   make([]interface{}, 0),
    }
}

func (s *SqlResult) Add(field string, record interface{}) {
    s.fields = append(s.fields, field)
    s.vals = append(s.vals, record)
}

func (s *SqlResult) ToMap() map[string]interface{} {
    results := make(map[string]interface{})
    for i, f := range s.fields {
        results[f] = s.vals[i]
    }
    return results
}

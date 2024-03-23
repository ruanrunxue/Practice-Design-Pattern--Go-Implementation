> 上一篇：[【Go实现】实践GoF的23种设计模式：适配器模式](https://mp.weixin.qq.com/s/dfVmdMQmIErGDBNb8RM-TQ)
>
> **简单的分布式应用系统**（示例代码工程）：[https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation)

## 简介

**解释器模式**（Interpreter Pattern）应该是 GoF 的 23 种设计模式中使用频率最少的一种了，它的应用场景较为局限。

GoF 对它的定义如下：

> Given a language, define a represention for its grammar along with an interpreter that uses the representation to interpret sentences in the language.

从定义可以看出，解释器模式主要运用于**简单的语法解析**场景，比如简单的领域特定语言（DSL）。举个例子，我们可以使用解析器模式来对“1+2+3-4+1”这样的文本表达式完成解析，并得到最终答案“3”。

**解释器模式的整体思想是分而治之**，每一个语法规则都使用一个类或者结构体（我们称之为 Rule Struct）来定义，它们相互独立，比如前一个例子中，“+” 和 “-” 都各自定义为一个 Rule Struct。因此，解释器模式的可扩展性很好。

通常，我们还能使用**抽象语法树**（Abstract Syntax Tree，AST）来直观地表示待解释的表达式，比如“1+2+3-4+1”可以表示成这样：

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2024-03-18-001951.png)

## UML 结构

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2024-03-17-234132.png)

解释器模式通常有 4 种角色：

- **Context**：解释上下文，包含了解释语法需要的所有信息，它是的生命周期贯穿整个解释过程，是一个全局对象。
- **AbstractExpression**：声明了解释语法的方法，通常只有 `Interpret(*Context)` 一个方法。
- **TerminalExpression**：实现了 AbstractExpression 接口，定义了**终结表达式**的解析逻辑。终结表达式在抽象语法树中作为叶子节点。
- **NonterminalExpression**：实现了 AbstractExpression 接口，定义了**非终结表达式**的解析逻辑。在抽象语法树中，除了叶子节点，其他节点都是非终结表达式。NonterminalExpression 通常会比 TerminalExpression 更复杂一些。

## 场景上下文

在 [简单的分布式应用系统](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation/blob/main/docs/go_ractice_design_pattern__solid_principle.md)（示例代码工程）中，db 模块用来存储服务注册信息和系统监控数据，它是一个 key-value 数据库。为了更高的易用性，它支持简单的 SQL 查询功能。用户在终端控制台上可以通过 SQL 语句来查询数据库中的数据：

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2023-12-03-023656.png)

简单起见，我们实现的 SQL 固定为 `select xxx,xxx,xxx from xxx where xxx=xxx;`  的形式，为此，我们要实现 3 个 TerminalExpression，即 `SelectExpression`、`FromExpression` 和 `WhereExpression`，分别解释 select 语句、from 语句、where 语句；以及 1 个 NonterminalExpression，即 `CompoundExpression`，用来解释整个 SQL 语句。

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2024-03-18-234152.png)

##  代码实现

```go
// demo/db/sql.go
package db

// 关键点1：定义Context结构体/类，这里是SqlContext，里面存放解析过程所需的状态和数据，以及结果数据
// SqlContext SQL解析器上下文，保存各个表达式解析的中间结果
// 当前只支持基于主键的查询SQL语句
type SqlContext struct {
    tableName  string
    fields     []string
    primaryKey interface{}
}

...

// 关键点2：定义AbstractExpression接口，这里是SqlExpression，其中Interpret方法以Context作为入参
// SqlExpression Sql表达式抽象接口，每个词、符号和句子都属于表达式
type SqlExpression interface {
    Interpret(ctx *SqlContext) error
}

// 关键点3：定义TerminalExpression，实现AbstractExpression接口，这里是SelectExpression、FromExpression和WhereExpression
// SelectExpression select语句解析逻辑，select关键字后面跟的为field，以,分割，比如select Id,name
type SelectExpression struct {
    fields string
}

func (s *SelectExpression) Interpret(ctx *SqlContext) error {
    fields := strings.Split(s.fields, ",")
    if len(fields) == 0 {
        return ErrSqlInvalidGrammar
    }
    // 关键点4：在解析过程中将状态或者结果数据存储到Context里面
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

// WhereExpression where语句解析逻辑，where关键字后面跟的是主键过滤条件，比如where id='1'
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

// 关键点5：实现NonterminalExpression，这里是CompoundExpression，它在解释过程中会引用到TerminalExpression，可以将TerminalExpression作为成员变量，也可以在Interpret方法中直接创建新对象。
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
    // 关键点6：在NonterminalExpression的Interpret方法中，调用TerminalExpression的Interpret方法完成对语句的解释。
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
```

客户端这么使用：

```go
// demo/db/memory_db.go
package db

// memoryDb 内存数据库
type memoryDb struct {
	tables sync.Map // key为tableName，value为table
}

...

func (m *memoryDb) ExecSql(sql string) (*SqlResult, error) {
	ctx := NewSqlContext()
	express := &CompoundExpression{sql: sql}
	if err := express.Interpret(ctx); err != nil {
		return nil, ErrSqlInvalidGrammar
	}
  // 关键点7：解释成功后，从Context中获取解释结果信息
	table, ok := m.tables.Load(ctx.TableName())
	if !ok {
		return nil, ErrTableNotExist
	}
	record, ok := table.(*Table).records[ctx.PrimaryKey()]
	if !ok {
		return nil, ErrRecordNotFound
	}
	result := NewSqlResult()
	for _, f := range ctx.Fields() {
		field := strings.ToLower(f)
		if idx, ok := table.(*Table).metadata[field]; ok {
			result.Add(field, record.values[idx])
		}
	}
	return result, nil
}
```

总结实现解释器模式的几个关键点：

1. 定义 Context 结构体/类，这里是 `SqlContext`，里面存放解释过程所需的状态和数据，也会存储解释结果。
2. 定义 AbstractExpression 接口，这里是 `SqlExpression`，其中 `Interpret` 方法以 Context 作为入参。
3. 定义 TerminalExpression 结构体，并实现 AbstractExpression 接口，这里是 `SelectExpression`、`FromExpression` 和 `WhereExpression`。
4. 将 `Interpret` 方法解释过程中产生的过程状态、数据存储在 Context 上，使得其他 Expression 在解释过程中能够访问。
5. 实现 NonterminalExpression，这里是 `CompoundExpression`，它在解释过程中会引用到 TerminalExpression，可以把 TerminalExpression 作为成员变量，也可以在 Interpret 方法中直接创建新对象。
6. 在 NonterminalExpression 的 Interpret 方法中，调用 TerminalExpression 的 Interpret 方法完成对语句的解释。这里是 `CompoundExpression.Interpret` 调用 `SelectExpression.Interpret`、`FromExpression.Interpret` 和 `WhereExpression.Interpret` 完成对 SQL 的解释。
7. 解释成功后，从 Context 中获取解释结果。

##  扩展

### 领域特定语言 DSL

在前文介绍解释器模式时有提到，它常用于对领域特定语言 DSL 的解释场景，那么什么是 DSL 呢？下面我们将简单介绍一下。

维基百科对 DSL 的定义如下：

> A **domain-specific language** (**DSL**) is a computer language specialized to a particular application domain. 

可见，DSL 是针对特定领域的一种计算机语言，与之相对的是 GPL，General Purpose Language，即通用编程语言。我们常用的 C/C++，Java，Go 等都属于 GPL 的范畴。

DSL 又可细分成 2 类：

- **External DSL** ：此类 DSL 拥有独立的语法以及解释器，比如 CSS 用于定义 Web 网页的样式和布局、SQL 用于数据查询、XML 和 YAML 用于配置管理，它们都是典型的 External DSL。

  ```sql
  # External DSL举例，SQL
  select id,name from regions where id=‘1’;
  ```

-  **Internal DSL**：此类 DSL 构建与 GPL 之上，比如流式接口 fluent interface、单元测试中的 Mock 库，它们可以提升 GPL 的易用性和易理解性。

  ```java
  // Internal DSL，Java中的Mockito库
  Mockito.when(mockDemo.isTrue()).thenReturn(1);
  ```

Martin Fowler 大神专门写了一本书《[领域特定语言](https://martinfowler.com/books/dsl.html)》来介绍 DSL，更多详细、专业的知识请移步这里。

## 典型应用场景

- **简单的语法解析**。解释器模式的运用场景较为单一，主要运用于**简单的语法解析**场景，比如简单的领域特定语言（DSL）。

## 优缺点

### 优点

- **易于扩展**。前文提到，使用解释器模式进行语法解释时，每种语法规则都会有对应的 Expression 结构体/类。因此，新增一种语法规则会非常的容易；类似地，改变一种已有的语法规则的解释方式也是很容易，单点改动即可。

### 缺点

- **不适用于复杂的语法解释**。当语法过于复杂时，Expression 结构体/类的数量将会变得很多，从而难以维护。


## 与其他模式的关联

解释器模式通常与**组合模式**（Composite Pattern）结合在一起使用，UML 结构图中的 NonterminalExpression 和 AbstractExpression 的就是组合关系。

另外，解释器模式这种分而治之的方法，与**状态模式**（State Pattern）中每种状态处理各种的逻辑很是类似。

### 文章配图

可以在 [用Keynote画出手绘风格的配图](https://mp.weixin.qq.com/s/-sYW-oa6KzTR9LNdMWCSnQ) 中找到文章的绘图方法。

> #### 参考
>
> [1] [【Go实现】实践GoF的23种设计模式：SOLID原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA), 元闰子
>
> [2] [Design Patterns, Chapter 5. Behavioral Patterns](https://learning.oreilly.com/library/view/design-patterns-elements/0201633612/ch05.html), GoF
>
> [3] [Domain-Specific Languages Guide](https://martinfowler.com/dsl.html), Martin Fowler
>
> 更多文章请关注微信公众号：**元闰子的邀请**

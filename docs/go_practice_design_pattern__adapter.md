> 上一篇：[【Go实现】实践GoF的23种设计模式：备忘录模式](https://mp.weixin.qq.com/s/1ZPv2pk_b8iivOIebX0-dQ)
>
> **简单的分布式应用系统**（示例代码工程）：[https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation)

## 简介

**适配器模式**（Adapter）是最常用的结构型模式之一，在现实生活中，适配器模式也是处处可见，比如电源插头转换器，它可以让英式的插头工作在中式的插座上。

GoF 对它的定义如下：

> Convert the interface of a class into another interface clients expect. Adapter lets classes work together that couldn’t otherwise because of incompatible interfaces.

简单来说，就是**适配器模式让原本因为接口不匹配而无法一起工作的两个类/结构体能够一起工作**。

适配器模式所做的就是**将一个接口 `Adaptee`，通过适配器 `Adapter` 转换成 Client 所期望的另一个接口 `Target` 来使用**，实现原理也很简单，就是 `Adapter` 通过实现 `Target` 接口，并在对应的方法中调用 `Adaptee` 的接口实现。

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2023-11-29-235329.png)

## UML 结构

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2023-12-02-022240.png)

## 场景上下文

在 [简单的分布式应用系统](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation/blob/main/docs/go_ractice_design_pattern__solid_principle.md)（示例代码工程）中，db 模块用来存储服务注册信息和系统监控数据，它是一个 key-value 数据库。在 [访问者模式](https://mp.weixin.qq.com/s/0qAIhbh4g_9Cde5tR2x24A) 中，我们为它实现了 Table 的按列查询功能；同时，我们也为它实现了简单的 SQL 查询功能（将会在 **解释器模式** 中介绍），查询的结果是 `SqlResult` 结构体，它提供一个 `toMap` 方法将结果转换成 `map` 。

为了方便用户使用，我们将实现在终端控制台上提供人机交互的能力，如下所示，用户输入 SQL 语句，后台返回查询结果：

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2023-12-03-023656.png)

终端控制台的具体实现为 `Console`，为了提供可扩展的查询结果显示样式，我们设计了 `ConsoleRender` 接口，但因 `SqlResult` 并未实现该接口，所以 `Console` 无法直接渲染 `SqlResult` 的查询结果。

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2023-12-03-154134.png)

为此，我们需要实现一个适配器，让 `Console` 能够通过适配器将 `SqlResult` 的查询结果渲染出来。示例中，我们设计了适配器 `TableRender`，它实现了 `ConsoleRender` 接口，并以表格的形式渲染出查询结果，如前文所示。

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2023-12-03-155326.png)

##  代码实现

```go
// demo/db/sql.go
package db

// Adaptee SQL语句执行返回的结果，并未实现Target接口
type SqlResult struct {
    fields []string
    vals   []interface{}
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

// demo/db/console.go
package db

// Client 终端控制台
type Console struct {
    db Db
}

// Output 调用ConsoleRender完成对查询结果的渲染输出
func (c *Console) Output(render ConsoleRender) {
    fmt.Println(render.Render())
}

// Target接口，控制台db查询结果渲染接口
type ConsoleRender interface {
    Render() string
}

// TableRender表格形式的查询结果渲染Adapter
// 关键点1: 定义Adapter结构体/类
type TableRender struct {
    // 关键点2: 在Adapter中聚合Adaptee，这里是把SqlResult作为TableRender的成员变量
    result *SqlResult
}

// 关键点3: 实现Target接口，这里是实现了ConsoleRender接口
func (t *TableRender) Render() string {
    // 关键点4: 在Target接口实现中，调用Adaptee的原有方法实现具体的业务逻辑
    vals := t.result.ToMap()
    var header []string
    var data []string
    for key, val := range vals {
        header = append(header, key)
        data = append(data, fmt.Sprintf("%v", val))
    }
    builder := &strings.Builder{}
    table := tablewriter.NewWriter(builder)
    table.SetHeader(header)
    table.Append(data)
    table.Render()
    return builder.String()
}

// 这里是另一个Adapter，实现了将error渲染的功能
type ErrorRender struct {
    err error
}

func (e *ErrorRender) Render() string {
    return e.err.Error()
}

```

客户端这么使用：

```go
func (c *Console) Start() {
    fmt.Println("welcome to Demo DB, enter exit to end!")
    fmt.Println("> please enter a sql expression:")
    fmt.Print("> ")
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        sql := scanner.Text()
        if sql == "exit" {
            break
        }
        result, err := c.db.ExecSql(sql)
        if err == nil {
            // 关键点5：在需要Target接口的地方，传入适配器Adapter实例，其中创建Adapter实例时需要传入Adaptee实例
            c.Output(NewTableRender(result))
        } else {
            c.Output(NewErrorRender(err))
        }
        fmt.Println("> please enter a sql expression:")
        fmt.Print("> ")
    }
}
```

在已经有了 Target 接口（`ConsoleRender`）和 Adaptee（`SqlResult`）的前提下，总结实现适配器模式的几个关键点：

1. 定义 Adapter 结构体/类，这里是 `TableRender` 结构体。
2. 在 Adapter 中聚合 Adaptee，这里是把 `SqlResult` 作为 `TableRender` 的成员变量。
3. Adapter 实现 Target 接口，这里是 `TableRender`  实现了 `ConsoleRender` 接口。
4. 在 Target 接口实现中，调用 Adaptee 的原有方法实现具体的业务逻辑，这里是在 `TableRender.Render()` 调用 `SqlResult.ToMap()` 方法，得到查询结果，然后再对结果进行渲染。
5. 在 Client 需要 Target 接口的地方，传入适配器 Adapter 实例，其中创建 Adapter 实例时传入 Adaptee 实例。这里是在 `NewTableRender()` 创建 `TableRender` 实例时，传入 `SqlResult` 作为入参，随后将 `TableRender` 实例传入 `Console.Output()` 方法。

##  扩展

### 适配器模式在 Gin 中的运用

Gin 是一个高性能的 Web 框架，它的常见用法如下：

```go
// 用户自定义的请求处理函数，类型为gin.HandlerFunc
func myGinHandler(c *gin.Context) {
    ... // 具体处理请求的逻辑
}

func main() {
    // 创建默认的route引擎,类型为gin.Engine
    r := gin.Default()
    // route定义
    r.GET("/my-route", myGinHandler)
    // route引擎启动
    r.Run()
}
```

在实际运用场景中，可能存在这种情况。用户起初的 Web 框架使用了 Go 原生的 `net/http`，使用场景如下：

```go
// 用户自定义的请求处理函数，类型为http.Handler
func myHttpHandler(w http.ResponseWriter, r *http.Request) {
    ... // 具体处理请求的逻辑
}

func main() {
    // route定义
    http.HandleFunc("/my-route", myHttpHandler)
    // route启动
    http.ListenAndServe(":8080", nil)
}
```

因性能问题，当前客户准备切换至 Gin 框架，显然，`myHttpHandler` 因接口不兼容，不能直接注册到 `gin.Default()` 上。为了方便用户，Gin 框架提供了一个适配器 `gin.WrapH`，可以将 `http.Handler` 类型转换成 `gin.HandlerFunc` 类型，它的定义如下：

```go
// WrapH is a helper function for wrapping http.Handler and returns a Gin middleware.
func WrapH(h http.Handler) HandlerFunc {
	  return func(c *Context) {
		  h.ServeHTTP(c.Writer, c.Request)
	  }
}
```

使用方法如下：

```go
// 用户自定义的请求处理函数，类型为http.Handler
func myHttpHandler(w http.ResponseWriter, r *http.Request) {
    ... // 具体处理请求的逻辑
}

func main() {
    // 创建默认的route引擎
    r := gin.Default()
    // route定义
    r.GET("/my-route", gin.WrapH(myHttpHandler))
    // route引擎启动
    r.Run()
}
```

在这个例子中，`gin.Engine` 就是 Client，`gin.HandlerFunc` 是 Target 接口，`http.Handler` 是 Adaptee，`gin.WrapH` 是 Adapter。这是一个 Go 风格的适配器模式实现，以更为简洁的 `func` 替代了 `struct`。

## 典型应用场景

- 将一个接口 A 转换成用户希望的另外一个接口 B，这样就能使原来不兼容的接口 A 和接口 B 相互协作。
- 老系统的重构。在不改变原有接口的情况下，让老接口适配到新的接口。

## 优缺点

### 优点

1. 能够使 Adaptee 和 Target 之间解耦。通过引入新的 Adapter 来适配 Target，Adaptee 无须修改，符合[开闭原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA)。
2. 灵活性好，能够很方便地通过不同的适配器来适配不同的接口。

### 缺点

1. 增加代码复杂度。适配器模式需要新增适配器，如果滥用会导致系统的代码复杂度增大。


## 与其他模式的关联

适配器模式 和 [装饰者模式](https://mp.weixin.qq.com/s/NT6_KOY_hGkA-y2b4fw45A)、[代理模式](https://mp.weixin.qq.com/s/_Z86PUn6hHgXh_4HWGv0Bg) 在 UML 结构上具有一定的相似性。但适配器模式改变原有对象的接口，但不改变原有功能；而装饰者模式和代理模式则在不改变接口的情况下，增强原有对象的功能。

### 文章配图

可以在 [用Keynote画出手绘风格的配图](https://mp.weixin.qq.com/s/-sYW-oa6KzTR9LNdMWCSnQ) 中找到文章的绘图方法。

> #### 参考
>
> [1] [【Go实现】实践GoF的23种设计模式：SOLID原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA), 元闰子
>
> [2] [Design Patterns, Chapter 4. Structural Patterns](https://learning.oreilly.com/library/view/design-patterns-elements/0201633612/ch04.html), GoF
>
> [3] [适配器模式](https://refactoringguru.cn/design-patterns/adapter), refactoringguru.cn
>
> [4] [Gin Web Framework](https://github.com/gin-gonic/gin), Gin
>
> 更多文章请关注微信公众号：**元闰子的邀请**

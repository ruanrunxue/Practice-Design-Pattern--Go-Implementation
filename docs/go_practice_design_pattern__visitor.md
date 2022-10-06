> 上一篇：[【Go实现】实践GoF的23种设计模式：迭代器模式](https://mp.weixin.qq.com/s/IFVH7VGaQQGmgr2nta7Pww)
>
> **简单的分布式应用系统**（示例代码工程）：[https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation)

## 简介

GoF 对**访问者模式**（**Visitor Pattern**）的定义如下：

> Represent an operation to be performed on the elements of an object structure. Visitor lets you define a new operation without changing the classes of the elements on which it operates.

访问者模式的目的是，**解耦数据结构和算法**，使得系统能够在不改变现有代码结构的基础上，为对象新增一种新的操作。

上一篇介绍的 [迭代器模式](https://mp.weixin.qq.com/s/IFVH7VGaQQGmgr2nta7Pww) 也做到了数据结构和算法的解耦，不过它专注于遍历算法。访问者模式，则在遍历的同时，将**操作**作用到数据结构上，一个常见的应用场景是语法树的解析。

## UML 结构

![](https://tva1.sinaimg.cn/large/006y8mN6gy1h6ve9yoetwj318y0q4jwu.jpg)

## 场景上下文

在 [简单的分布式应用系统](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation/blob/main/docs/go_ractice_design_pattern__solid_principle.md)（示例代码工程）中，db 模块用来存储服务注册和监控信息，它是一个 key-value 数据库。另外，我们给 db 模块抽象出 `Table` 对象：

```go
// demo/db/table.go
package db

// Table 数据表定义
type Table struct {
    name            string
    metadata        map[string]int // key为属性名，value属性值的索引, 对应到record上存储
    records         map[interface{}]record
    iteratorFactory TableIteratorFactory // 默认使用随机迭代器
}
```

目的是提供类似于关系型数据库的**按列查询**能力，比如：

![](https://tva1.sinaimg.cn/large/006y8mN6gy1h6vf8t4xekj31600keads.jpg)

上述的按列查询只是等值比较，未来还可能会实现正则表达式匹配等方式，因此我们需要设计出可供未来扩展的接口。这种场景，使用访问者模式正合适。

## 代码实现

```go
// demo/db/table_visitor.go
package db

// 关键点1: 定义表查询的访问者抽象接口，允许后续扩展查询方式
type TableVisitor interface {
    // 关键点2: Visit方法以Element作为入参，这里的Element为Table对象
    Visit(table *Table) ([]interface{}, error)
}

// 关键点3: 定义Visitor抽象接口的实现对象，这里FieldEqVisitor实现按列等值查询逻辑
type FieldEqVisitor struct {
    field string
    value interface{}
}

// 关键点4: 为FieldEqVisitor定义Visit方法，实现具体的等值查询逻辑
func (f *FieldEqVisitor) Visit(table *Table) ([]interface{}, error) {
    result := make([]interface{}, 0)
    idx, ok := table.metadata[f.field]
    if !ok {
        return nil, ErrRecordNotFound
    }
    for _, r := range table.records {
        if reflect.DeepEqual(r.values[idx], f.value) {
            result = append(result, r)
        }
    }
    if len(result) == 0 {
        return nil, ErrRecordNotFound
    }
    return result, nil
}

func NewFieldEqVisitor(field string, value interface{}) *FieldEqVisitor {
    return &FieldEqVisitor{
        field: field,
        value: value,
    }
}

// demo/db/table.go
package db

type Table struct {...}
// 关键点5: 为Element定义Accept方法，入参为Visitor接口
func (t *Table) Accept(visitor TableVisitor) ([]interface{}, error) {
    return visitor.Visit(t)
}
```

客户端可以这么使用：

```go
func client() {
    table := NewTable("testRegion").WithType(reflect.TypeOf(new(testRegion)))
    table.Insert(1, &testRegion{Id: 1, Name: "beijing"})
    table.Insert(2, &testRegion{Id: 2, Name: "beijing"})
    table.Insert(3, &testRegion{Id: 3, Name: "guangdong"})

  visitor := NewFieldEqVisitor("name", "beijing")
    result, err := table.Accept(visitor)
    if err != nil {
        t.Error(err)
    }
    if len(result) != 2 {
        t.Errorf("visit failed, want 2, got %d", len(result))
    }
}
```

总结实现访问者模式的几个关键点：

1. 定义访问者抽象接口，上述例子为 `TableVisitor`， 目的是允许后续扩展表查询方式。
2. 访问者抽象接口中，`Visit` 方法以 Element 作为入参，上述例子中， Element 为 `Table` 对象。
3. 为 Visitor 抽象接口定义具体的实现对象，上述例子为 `FieldEqVisitor`。
4. 在访问者的 `Visit` 方法中实现具体的业务逻辑，上述例子中 `FieldEqVisitor.Visit(...)` 实现了按列等值查询逻辑。
5. 在被访问者 Element 中定义 Accept 方法，以访问者 Visitor 作为入参。上述例子中为 `Table.Accept(...)` 方法。

## 扩展

### Go 风格实现

上述实现是典型的面向对象风格，下面以 Go 风格重新实现访问者模式：

```go
// demo/db/table_visitor_func.go
package db

// 关键点1: 定义一个访问者函数类型
type TableVisitorFunc func(table *Table) ([]interface{}, error)

// 关键点2: 定义工厂方法，工厂方法返回的是一个访问者函数，实现了具体的访问逻辑
func NewFieldEqVisitorFunc(field string, value interface{}) TableVisitorFunc {
    return func(table *Table) ([]interface{}, error) {
        result := make([]interface{}, 0)
        idx, ok := table.metadata[field]
        if !ok {
            return nil, ErrRecordNotFound
        }
        for _, r := range table.records {
            if reflect.DeepEqual(r.values[idx], value) {
                result = append(result, r)
            }
        }
        if len(result) == 0 {
            return nil, ErrRecordNotFound
        }
        return result, nil
    }
}

// 关键点3: 为Element定义Accept方法，入参为Visitor函数类型
func (t *Table) AcceptFunc(visitorFunc TableVisitorFunc) ([]interface{}, error) {
    return visitorFunc(t)
}
```

客户端可以这么使用：

```go
func client() {
    table := NewTable("testRegion").WithType(reflect.TypeOf(new(testRegion)))
    table.Insert(1, &testRegion{Id: 1, Name: "beijing"})
    table.Insert(2, &testRegion{Id: 2, Name: "beijing"})
    table.Insert(3, &testRegion{Id: 3, Name: "guangdong"})

    result, err := table.AcceptFunc(NewFieldEqVisitorFunc("name", "beijing"))
    if err != nil {
        t.Error(err)
    }
    if len(result) != 2 {
        t.Errorf("visit failed, want 2, got %d", len(result))
    }
}
```

Go 风格的实现，利用了函数闭包的特点，更加简洁了。

总结几个实现关键点：

1. 定义一个访问者函数类型，函数签名以 Element 作为入参，上述例子为 `TableVisitorFunc` 类型。
2. 定义一个工厂方法，工厂方法返回的是具体的访问访问者函数，上述例子为 `NewFieldEqVisitorFunc` 方法。这里利用了函数闭包的特性，在访问者函数中直接引用工厂方法的入参，与 `FieldEqVisitor` 中持有两个成员属性的效果一样。
3. 为 Element 定义 Accept 方法，入参为 Visitor 函数类型 ，上述例子是 `Table.AcceptFunc(...)` 方法。

### 与迭代器模式结合

**访问者模式经常与迭代器模式一起使用**。比如上述例子中，如果你定义的 Visitor 实现不在 db 包内，那么就无法直接访问 `Table` 的数据，这时就需要通过 `Table` 提供的迭代器来实现。

在 [简单的分布式应用系统](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation/blob/main/docs/go_ractice_design_pattern__solid_principle.md)（示例代码工程）中，db 模块存储的服务注册信息如下：

```go
// demo/service/registry/model/service_profile.go
package model

// ServiceProfileRecord 存储在数据库里的类型
type ServiceProfileRecord struct {
    Id       string        // 服务ID
    Type     ServiceType   // 服务类型
    Status   ServiceStatus // 服务状态
    Ip       string        // 服务IP
    Port     int           // 服务端口
    RegionId string        // 服务所属regionId
    Priority int           // 服务优先级，范围0～100，值越低，优先级越高
    Load     int           // 服务负载，负载越高表示服务处理的业务压力越大
}
```

现在，我们要查询符合指定 `ServiceId` 和 `ServiceType` 的服务记录，可以这么实现一个 Visitor：

```go
// demo/service/registry/model/service_profile.go
package model

type ServiceProfileVisitor struct {
    svcId   string
    svcType ServiceType
}

func (s *ServiceProfileVisitor) Visit(table *db.Table) ([]interface{}, error) {
    var result []interface{}
  // 通过迭代器来遍历Table的所有数据
    iter := table.Iterator()
    for iter.HasNext() {
        profile := new(ServiceProfileRecord)
        if err := iter.Next(profile); err != nil {
            return nil, err
        }
        // 先匹配ServiceId，如果一致则无须匹配ServiceType
        if profile.Id != "" && profile.Id == s.svcId {
            result = append(result, profile)
            continue
        }
        // ServiceId匹配不上，再匹配ServiceType
        if profile.Type != "" && profile.Type == s.svcType {
            result = append(result, profile)
        }
    }
    return result, nil
}
```

## 典型应用场景

- k8s 中，kubectl 通过访问者模式来处理用户定义的各类资源。

- 编译器中，通常使用访问者模式来实现对语法树解析，比如 LLVM。
- 希望对一个复杂的数据结构执行某些操作，并支持后续扩展。

## 优缺点

### 优点

- 数据结构和操作算法解耦，符合 [单一职责原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA)。
- 支持对数据结构扩展多种操作，具备较强的可扩展性，符合 [开闭原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA)。

### 缺点

- 访问者模式某种程度上，要求数据结构必须对外暴露其内在实现，否则访问者就无法遍历其中数据（可以结合迭代器模式来解决该问题）。
- 如果被访问对象内的数据结构变更，可能要更新所有的访问者实现。

## 与其他模式的关联

- 访问者模式 经常和 [迭代器模式](https://mp.weixin.qq.com/s/IFVH7VGaQQGmgr2nta7Pww) 一起使用，使得被访问对象无须向外暴露内在数据结构。
- 也经常和 **组合模式** 一起使用，比如在语法树解析中，递归访问和解析树的每个节点（节点组合成树）。

### 文章配图

可以在 [用Keynote画出手绘风格的配图](https://mp.weixin.qq.com/s/-sYW-oa6KzTR9LNdMWCSnQ) 中找到文章的绘图方法。

> #### 参考
>
> [1] [【Go实现】实践GoF的23种设计模式：SOLID原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA), 元闰子
>
> [2] [【Go实现】实践GoF的23种设计模式：迭代器模式](https://mp.weixin.qq.com/s/IFVH7VGaQQGmgr2nta7Pww), 元闰子
>
> [3] [Design Patterns, Chapter 5. Behavioral Patterns](https://learning.oreilly.com/library/view/design-patterns-elements/0201633612/), GoF
>
> [4] [GO 编程模式：K8S VISITOR 模式](https://coolshell.cn/articles/21263.html), 酷壳
>
> [5] [访问者模式](https://refactoringguru.cn/design-patterns/visitor), refactoringguru.cn
>
> 更多文章请关注微信公众号：**元闰子的邀请**

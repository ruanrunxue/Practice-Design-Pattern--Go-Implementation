> 上一篇：[【Go实现】实践GoF的23种设计模式：代理模式](https://mp.weixin.qq.com/s/_Z86PUn6hHgXh_4HWGv0Bg)
>
> **简单的分布式应用系统**（示例代码工程）：[https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation)

## 简介

现在的软件系统往往是分层设计。在业务层执行一次请求时，我们很清楚请求的上下文，包括，请求是做什么的、参数有哪些、请求的接收者是谁、返回值是怎样的。相反，基础设施层并不需要完全清楚业务上下文，它只需知道请求的接收者是谁即可，否则就耦合过深了。

因此，我们需要对请求进行抽象，将上下文信息封装到请求对象里，这其实就是命令模式，而该请求对象就是 Command。

GoF 对**命令模式**（Command Pattern）的定义如下：

> Encapsulate a request as an object, thereby letting you parameterize clients with different requests, queue or log requests, and support undoable operations.

也即，**命令模式可将请求转换为一个包含与请求相关的所有信息的对象， 它能将请求参数化、延迟执行、实现 Undo / Redo 操作等**。

上述的**请求**是广义上的概念，可以是网络请求，也可以是函数调用，更通用地，指一个**动作**。

命令模式主要包含 3 种角色：

1. **Command**，命令，是对请求的抽象。具体的命令实现时，通常会引用 Receiver。
2. **Invoker**，请求的发起发起方，它并不清楚 Command 和 Receiver 的实现细节，只管调用命令的接口。
3. **Receiver**，请求的接收方。

![](https://tva1.sinaimg.cn/large/008vxvgGgy1h97uo4ir27j30zo0e2409.jpg)

命令模式，一方面，能够使得 Invoker 与 Receiver 消除彼此之间的耦合，让对象之间的调用关系更加灵活；另一方面，能够很方便地实现延迟执行、Undo、Redo 等操作，因此被广泛应用在软件设计中。

## UML 结构

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2023-11-22-233054.png)

## 场景上下文

在 [简单的分布式应用系统](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation/blob/main/docs/go_ractice_design_pattern__solid_principle.md)（示例代码工程）中，db 模块用来存储服务注册信息和系统监控数据。其中，服务注册信息拆成了 `profiles` 和 `regions` 两个表，在服务发现的业务逻辑中，通常需要同时操作两个表，为了避免两个表数据不一致的问题，**db 模块需要提供事务功能**:

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2023-11-22-233012.png)

事务的核心功能之一是，当其中某个语句执行失败时，之前已执行成功的语句能够回滚，而使用命令模式能够很方便地实现该功能。

## 代码实现

```go
// demo/db/transaction.go
package db

// Command 执行数据库操作的命令接口
// 关键点1: 定义命令抽象接口
type Command interface {
  // 关键点2: 命令抽象接口中声明执行命令的方法
    Exec() error // Exec 执行insert、update、delete命令
  // 关键点3: 如果有撤销功能，则需要定义Undo方法
    Undo() // Undo 回滚命令
    setDb(db Db) // SetDb 设置关联的数据库
}

// Transaction Db事务实现，事务接口的调用顺序为begin -> exec -> exec > ... -> commit
// 关键点4: 定义Invoker对象
type Transaction struct {
    name string
    db   Db
  // 关键点5: Invoker对象持有Command的引用
    cmds []Command
}
// Begin 开启一个事务
func (t *Transaction) Begin() {
    t.cmds = make([]Command, 0)
}
// Exec 在事务中执行命令，先缓存到cmds队列中，等commit时再执行
func (t *Transaction) Exec(cmd Command) error {
    if t.cmds == nil {
        return ErrTransactionNotBegin
    }
    cmd.setDb(t.db)
    t.cmds = append(t.cmds, cmd)
    return nil
}
// Commit 提交事务，执行队列中的命令，如果有命令失败，则回滚后返回错误
// 关键点6: 为Invoker对象定义Call方法，在方法内调用Command的执行方法Exec
func (t *Transaction) Commit() error {
    history := &cmdHistory{history: make([]Command, 0, len(t.cmds))}
    for _, cmd := range t.cmds {
        if err := cmd.Exec(); err != nil {
            history.rollback()
            return err
        }
        history.add(cmd)
    }
    return nil
}
// cmdHistory 命令执行历史
type cmdHistory struct {
    history []Command
}
func (c *cmdHistory) add(cmd Command) {
    c.history = append(c.history, cmd)
}
// 关键点7: 在回滚方法中，调用已执行命令的Undo方法
func (c *cmdHistory) rollback() {
    for i := len(c.history) - 1; i >= 0; i-- {
        c.history[i].Undo()
    }
}

// InsertCmd 插入命令
// 关键点8: 定义具体的命令类，实现Command接口
type InsertCmd struct {
  // 关键点9: 命令通常持有接收者的引用，以便在执行方法中与接收者交互
    db         Db
    tableName  string
    primaryKey interface{}
    newRecord  interface{}
}
// 关键点10: 命令对象执行方法中，调用Receiver的Action方法，这里的Receiver为db对象，Action方法为Insert方法
func (i *InsertCmd) Exec() error {
    return i.db.Insert(i.tableName, i.primaryKey, i.newRecord)
}
func (i *InsertCmd) Undo() {
    i.db.Delete(i.tableName, i.primaryKey)
}
func (i *InsertCmd) setDb(db Db) {
    i.db = db
}

// UpdateCmd 更新命令
type UpdateCmd struct {...}
// DeleteCmd 删除命令
type DeleteCmd struct {...}

```

客户端可以这么使用：

```go
func client() {
    transaction := db.CreateTransaction("register" + profile.Id)
    transaction.Begin()
    rcmd := db.NewUpdateCmd(regionTable).WithPrimaryKey(profile.Region.Id).WithRecord(profile.Region)
    transaction.Exec(rcmd)
    pcmd := db.NewUpdateCmd(profileTable).WithPrimaryKey(profile.Id).WithRecord(profile.ToTableRecord())
    transaction.Exec(pcmd)
    if err := transaction.Commit(); err != nil {
        return ... 
    }
  return ...
}
```

总结实现命令模式的几个关键点：

1. 定义命令抽象接口，本例子中为 `Command` 接口。
2. 在命令抽象接口中声明执行命令的方法，本例子中为 `Exec` 方法。
3. 如果要实现撤销功能，还需要为命令对象定义 `Undo` 方法，在操作回滚时调用。
4. 定义 `Invoker` 对象，本例子中为 `Transaction` 对象。
5. 在 Invoker 对象持有 Command 的引用，本例子为 `Command` 的切片 `cmds`。
6. 为 Invoker 对象定义 Call 方法，用于执行具体的命令，在方法内调用 Command 的执行方法 ，本例子中为 `Transaction.Commit` 方法。
7. 如果要实现撤销功能，还要在回滚方法中，调用已执行命令的 `Undo` 方法，本例子中为 `cmdHistory.rollback` 方法。
8. 定义具体的命令类，实现 `Command` 接口，本例子中为 `InsertCmd`、`UpdateCmd`、`DeleteCmd`。
9. 命令通常持有接收者的引用，以便在执行方法中与接收者交互。本例子中，Receiver 为 `Db` 对象。
10. 最后，在命令对象执行方法中，调用 Receiver 的 Action 方法，本例子中， Receiver 的 Action 方法为 `db.Insert` 方法。

值得注意的是，本例子中 `Transaction` 对象在 `Transaction.Exec` 方法中只是将 `Command` 保存在队列中，只有当调用 `Transaction.Commit` 方法时才延迟执行相应的命令。

## 扩展

### `os/exec` 中的命令模式

Go 标准库的 `os/exec` 包也用到了命令模式。

```go
package main

import (
  "os/exec"
)

// 对应命令模式中的Invoker
func main() {
  cmd := exec.Command("sleep", "1")
  err := cmd.Run()
}

```

在上述例子中，我们通过 `exec.Command` 方法将一个 shell 命令转换成一个**命令对象** `exec.Cmd`，其中的 `Cmd.Run()` 方法即是命令执行方法；而 `main()` 函数，对应到命令模式中的 Invoker；Receiver 则是操作系统执行 shell 命令的具体进程，从 `exec.Cmd` 的源码中可以看到：

```go
// src/os/exec/exec.go
package exec

// 对应命令模式中的Command
type Cmd struct {
  ...
  // 对应命令模式中的Receiver
    Process *os.Process
  ...
}

// 对应命令模式中Command的执行方法
func (c *Cmd) Run() error {
    if err := c.Start(); err != nil {
        return err
    }
    return c.Wait()
}

func (c *Cmd) Start() error {
  ...
  // Command与Receiver的交互
  c.Process, err = os.StartProcess(c.Path, c.argv(), &os.ProcAttr{...})
  ...
}
```

![](https://tva1.sinaimg.cn/large/008vxvgGgy1h9bvbx3j2yj31dk0p8q8m.jpg)

### CQRS 架构

**CQRS 架构**，全称为 Command Query Responsibility Segregation，命令查询职责隔离架构。CQRS 架构是微服务架构模式中的一种，**它利用事件（命令）来维护从多个服务复制数据的只读视图**，通过读写分离思想，提升微服务架构下查询的性能。

![](https://tva1.sinaimg.cn/large/008vxvgGgy1h9cw06dkfvj30n90ctab7.jpg)

CQRS 架构可分为 **命令端** 和 **查询端**，其中命令端负责数据的更新；查询端负责数据的查询。命令端的写数据库在数据更新时，会向查询端的只读数据库发送一个同步数据的事件，保证数据的最终一致性。

其中的命令端，就使用到了命令模式的思想，**将数据更新请求封装成命令，异步更新到写数据库中**。

## 典型应用场景

- **事务模式**。事务模式下往往需要 Undo 操作，使用命令模式实现起来很方便。
- **远程执行**。Go 标准库下的 `exec.Cmd`、`http.Client` 都属于该类型，将请求封装成命令来执行。
- **CQRS 架构**。微服务架构模式中的一种，通过命令模式来实现数据的异步更新。
- **延迟执行**。当你希望一个操作能够延迟执行时，通常会将它封装成命令，然后放到一个队列中。

## 优缺点

### 优点

1. 符合**单一职责原则**。在命令模式下，每个命令都是职责单一、松耦合的；当然也可以通过组合的方式，将多个简单的命令组合成一个负责的命令。
2. 可以很方便地实现操作的延迟执行、回滚、重做等。
3. 在分布式架构下，命令模式能够方便地实现异步的数据更新、方法调用等，提升性能。

### 缺点

1. 命令模式下，调用往往是异步的，而**异步会导致系统变得复杂**，问题出现时不好定位解决。
2. 随着业务越来越复杂，命令对象也会增多，代码会变得更难维护。

## 与其他模式的关联

在实现 Undo/Redo 操作时，你通常需要同时使用 命令模式 和 **备忘录模式**。

另外，命令模式 也常常和 [观察者模式](https://mp.weixin.qq.com/s/RZqZmjWm_NMzgf8nUhv2Xg) 一起出现，比如在 CQRS 架构中，当命令端更新数据库后，写数据库就会通过事件将数据同步到读数据库上，这里就用到了 [观察者模式](https://mp.weixin.qq.com/s/RZqZmjWm_NMzgf8nUhv2Xg)。

### 文章配图

可以在 [用Keynote画出手绘风格的配图](https://mp.weixin.qq.com/s/-sYW-oa6KzTR9LNdMWCSnQ) 中找到文章的绘图方法。

> #### 参考
>
> [1] [【Go实现】实践GoF的23种设计模式：SOLID原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA), 元闰子
>
> [2] [【Go实现】实践GoF的23种设计模式：观察者模式](https://mp.weixin.qq.com/s/RZqZmjWm_NMzgf8nUhv2Xg), 元闰子
>
> [3] [Design Patterns, Chapter 5. Behavioral Patterns](https://learning.oreilly.com/library/view/design-patterns-elements/0201633612/ch05.html), GoF
>
> [4] [命令模式](https://refactoringguru.cn/design-patterns/command), refactoringguru.cn
>
> [5] [The command pattern in Go](https://rolandsdev.blog/posts/the-command-pattern-in-go/), rolandjitsu
>
> [6] [CQRS 模式](https://learn.microsoft.com/zh-cn/azure/architecture/patterns/cqrs), microsoft azure
>
> [7] [CQRS Design Pattern in Microservices Architectures](https://medium.com/design-microservices-architecture-with-patterns/cqrs-design-pattern-in-microservices-architectures-5d41e359768c), Mehmet Ozkaya
>
> 更多文章请关注微信公众号：**元闰子的邀请**

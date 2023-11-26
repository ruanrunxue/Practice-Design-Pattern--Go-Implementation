> 上一篇：[【Go实现】实践GoF的23种设计模式：命令模式](https://mp.weixin.qq.com/s/p5ZMohLxt3Niy8VtJH_1_A)
>
> **简单的分布式应用系统**（示例代码工程）：[https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation)

## 简介

相对于[代理模式](https://mp.weixin.qq.com/s/_Z86PUn6hHgXh_4HWGv0Bg)、[工厂模式](https://mp.weixin.qq.com/s/PwHc31ANLDVMNiagtqucZQ)等设计模式，**备忘录模式**（Memento）在我们日常开发中出镜率并不高，除了应用场景的限制之外，另一个原因，可能是备忘录模式 UML 结构的几个概念比较晦涩难懂，难以映射到代码实现中。比如 Originator（原发器）和 Caretaker（负责人），从字面上很难看出它们在模式中的职责。

但从定义来看，备忘录模式又是简单易懂的，GoF 对备忘录模式的定义如下：

> Without violating encapsulation, capture and externalize an object’s internal state so that the object can be restored to this state later.

也即，**在不破坏封装的前提下，捕获一个对象的内部状态，并在该对象之外进行保存，以便在未来将对象恢复到原先保存的状态**。

从定义上看，备忘录模式有几个关键点：*封装*、*保存*、*恢复*。

对状态的封装，主要是为了未来状态修改或扩展时，不会引发[霰弹式修改](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA)；保存和恢复则是备忘录模式的主要特点，能够对当前对象的状态进行保存，并能够在未来某一时刻恢复出来。

现在，在回过头来看备忘录模式的 3 个角色就比较好理解了：

- **Memento**（备忘录）：是对状态的封装，可以是 `struct` ，也可以是 `interface`。
- **Originator**（原发器）：备忘录的创建者，备忘录里存储的就是 Originator 的状态。
- **Caretaker**（负责人）：负责对备忘录的保存和恢复，无须知道备忘录中的实现细节。

## UML 结构

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2023-11-22-232135.png)

## 场景上下文

在前文 [【Go实现】实践GoF的23种设计模式：命令模式](https://mp.weixin.qq.com/s/p5ZMohLxt3Niy8VtJH_1_A) 我们提到，在 [简单的分布式应用系统](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation/blob/main/docs/go_ractice_design_pattern__solid_principle.md)（示例代码工程）中，db 模块用来存储服务注册信息和系统监控数据。其中，服务注册信息拆成了 `profiles` 和 `regions` 两个表，在服务发现的业务逻辑中，通常需要同时操作两个表，为了避免两个表数据不一致的问题，**db 模块需要提供事务功能**:

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2023-11-22-233012.png)

事务的核心功能之一是，**当其中某个语句执行失败时，之前已执行成功的语句能够回滚**，前文我们已经介绍如何基于 [命令模式](https://mp.weixin.qq.com/s/p5ZMohLxt3Niy8VtJH_1_A) 搭建事务框架，下面我们将重点介绍，如何基于备忘录模式实现失败回滚的功能。

## 代码实现

```go
// demo/db/transaction.go
package db

// Command 执行数据库操作的命令接口，同时也是备忘录接口
// 关键点1：定义Memento接口，其中Exec方法相当于UML图中的SetState方法，调用后会将状态保存至Db中
type Command interface {
    Exec() error // Exec 执行insert、update、delete命令
    Undo() // Undo 回滚命令
    setDb(db Db) // SetDb 设置关联的数据库
}

// 关键点2：定义Originator，在本例子中，状态都是存储在Db对象中
type Db interface {...}

// Transaction Db事务实现，事务接口的调用顺序为begin -> exec -> exec > ... -> commit
// 关键点3：定义Caretaker，Transaction里实现了对语句的执行（Do）和回滚（Undo）操作
type Transaction struct {
    name string
    // 关键点4：在Caretaker（Transaction）中引用Originator（Db）对象，用于后续对其状态的保存和恢复
    db   Db
    // 注意，这里的cmds并非备忘录列表，真正的history在Commit方法中
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
func (t *Transaction) Commit() error {
    // 关键点5：定义备忘录列表，用于保存某一时刻的系统状态
    history := &cmdHistory{history: make([]Command, 0, len(t.cmds))}
    for _, cmd := range t.cmds {
        // 关键点6：执行Do方法
        if err := cmd.Exec(); err != nil {
            // 关键点8：当Do方法执行失败时，则进行Undo操作，根据备忘录history中的状态进行回滚
            history.rollback()
            return err
        }
        // 关键点7：如果Do方法执行成功，则将状态（cmd）保存在备忘录history中
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

func (c *cmdHistory) rollback() {
    for i := len(c.history) - 1; i >= 0; i-- {
        c.history[i].Undo()
    }
}

// InsertCmd 插入命令
// 关键点9: 定义具体的备忘录类，实现Memento接口
type InsertCmd struct {
    db         Db
    tableName  string
    primaryKey interface{}
    newRecord  interface{}
}

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

这里并没有完全按照标准的备忘录模式 UML 进行实现，但本质是一样的，总结起来有以下几个关键点：

1. 定义抽象备忘录 Memento 接口，这里为 `Command` 接口。`Command` 的实现是具体的数据库执行操作，并且存有对应的回滚操作，比如 `InsertCmd` 为“插入”操作，其对应的回滚操作为“删除”，我们保存的状态就是“删除”这一回滚操作。
2. 定义 Originator 结构体/接口，这里为 `Db` 接口。备忘录 `Command` 记录的就是它的状态。
3. 定义 Caretaker 结构体/接口，这里为 `Transaction` 结构体。`Transaction` 采用了延迟执行的设计，当调用 `Exec` 方法时只会将命令缓存到 `cmds` 队列中，等到调用 `Commit` 方法时才会执行。
4. 在 Caretaker 中引用 Originator 对象，用于后续对其状态的保存和恢复。这里为 `Transaction` 聚合了 `Db`。
5. 在 Caretaker 中定义备忘录列表，用于保存某一时刻的系统状态。这里为在 `Transaction.Commit` 方法中定义了 `cmdHistory` 对象，保存一直执行成功的 `Command`。
6. 执行 Caretaker 具体的业务逻辑，这里为在 `Transaction.Commit` 中调用 `Command.Exec` 方法，执行具体的数据库操作命令。
7. 业务逻辑执行成功后，保存当前的状态。这里为调用 `cmdHistory.add` 方法将 `Command` 保存起来。
8. 如果业务逻辑执行失败，则恢复到原来的状态。这里为调用`cmdHistory.rollback` 方法，反向执行已执行成功的 `Command` 的 `Undo` 方法进行状态恢复。
9. 根据具体的业务需要，定义具体的备忘录，这里定义了`InsertCmd` 、`UpdateCmd` 和 `DeleteCmd` 。

##  扩展

### MySQL 的 undo log 机制

MySQL 的 **undo log（回滚日志）机制**本质上用的就是备忘录模式的思想，前文中 `Transaction` 回滚机制实现的方法参考的就是 undo log 机制。

undo log 原理是，在提交事务之前，会把该事务对应的**回滚操作**（状态）先**保存**到 undo log 中，然后再提交事务，当出错的时候 MySQL 就可以利用 undo log 来回滚事务，即**恢复**原先的记录值。

比如，执行一条插入语句：

```sql
insert into region(id, name) values (1, "beijing");
```

那么，写入到 undo log 中对应的回滚语句为：

```sql
delete from region where id = 1;
```

当执行一条语句失败，需要回滚时，MySQL 就会从读取对应的回滚语句来执行，从而将数据恢复至事务提交之前的状态。undo log 是 MySQL 实现事务回滚和多版本控制（MVCC）的根基。

## 典型应用场景

- **事务回滚**。事务回滚的一种常见实现方法是 undo log，其本质上用的就是备忘录模式。
- **系统快照（Snapshot）**。多版本控制的用法，保存某一时刻的系统状态快照，以便在将来能够恢复。
- **撤销功能**。比如 Microsoft Offices 这类的文档编辑软件的撤销功能。

## 优缺点

### 优点

1. **提供了一种状态恢复的机制**，让系统能够方便地回到某个特定状态下。
1. **实现了对状态的封装**，能够在不破坏封装的前提下实现状态的保存和恢复。

### 缺点

1. **资源消耗大**。系统状态的保存意味着存储空间的消耗，本质上是空间换时间的策略。*undo log 是一种折中方案*，保存的状态并非某一时刻数据库的所有数据，而是一条反操作的 SQL 语句，存储空间大大减少。
1. **并发安全**。在多线程场景，实现备忘录模式时，要注意在保证状态的不变性，否则可能会有并发安全问题。

## 与其他模式的关联

在实现 Undo/Redo 操作时，你通常需要同时使用 **备忘录模式** 与 [命令模式](https://mp.weixin.qq.com/s/p5ZMohLxt3Niy8VtJH_1_A)。

另外，当你需要遍历备忘录对象中的成员时，通常会使用 [迭代器模式](https://mp.weixin.qq.com/s/IFVH7VGaQQGmgr2nta7Pww)，以防破坏对象的封装。

### 文章配图

可以在 [用Keynote画出手绘风格的配图](https://mp.weixin.qq.com/s/-sYW-oa6KzTR9LNdMWCSnQ) 中找到文章的绘图方法。

> #### 参考
>
> [1] [【Go实现】实践GoF的23种设计模式：SOLID原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA), 元闰子
>
> [2] [【Go实现】实践GoF的23种设计模式：命令模式](https://mp.weixin.qq.com/s/p5ZMohLxt3Niy8VtJH_1_A), 元闰子
>
> [3] [Design Patterns, Chapter 5. Behavioral Patterns](https://learning.oreilly.com/library/view/design-patterns-elements/0201633612/ch05.html), GoF
>
> [4] [备忘录模式](https://refactoringguru.cn/design-patterns/memento), refactoringguru.cn
>
> [5] [MySQL 8.0 Reference Manual :: 15.6.6 Undo Logs](https://dev.mysql.com/doc/refman/8.0/en/innodb-undo-logs.html), MySQL
>
> 更多文章请关注微信公众号：**元闰子的邀请**

> 上一篇：[【Go实现】实践GoF的23种设计模式：解释器模式](https://mp.weixin.qq.com/s/5Ttr-QKMoaV0GTlR_L8JbA)
>
> **简单的分布式应用系统**（示例代码工程）：[https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation)

## 简介

GoF 对**桥接模式**（Bridge Pattern）的定义如下：

> Decouple an abstraction from its implementation so that the two can vary independently.

也即，**将抽象部分和实现部分进行解耦，使得它们能够各自往独立的方向变化**。

桥接模式解决了在模块有多种变化方向的情况下，用继承所导致的类爆炸问题。

举个例子，一个产品有形状和颜色两个特征（变化方向），其中形状分为方形和圆形，颜色分为红色和蓝色。如果采用继承的设计方案，那么就需要新增4个产品子类：方形红色、圆形红色、方形蓝色、圆形红色。如果形状总共有 m 种变化，颜色有 n 种变化，那么就需要新增 m * n 个产品子类！

现在我们使用桥接模式进行优化，将形状和颜色分别设计为抽象接口独立出来，这样需要新增 2 个形状子类：方形和圆形，以及 2 个颜色子类：红色和蓝色。同样，如果形状总共有 m 种变化，颜色有 n 种变化，总共只需要新增 m + n 个子类！

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2024-04-01-235051.png)

上述例子中，我们通过将形状和颜色抽象为一个接口，使产品不再依赖于具体的形状和颜色细节，从而达到了解耦的目的。**桥接模式本质上就是面向接口编程，可以给系统带来很好的灵活性和可扩展性**。如果一个对象存在多个变化的方向，而且每个变化方向都需要扩展，那么使用桥接模式进行设计那是再合适不过了。

当然，Go 语言从语言特性本身就把继承剔除，但桥接模式中**分离变化、面向接口编程的思想**仍然值得学习。

## UML 结构

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2024-04-06-010127.png)

## 场景上下文

在 [简单的分布式应用系统](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation/blob/main/docs/go_ractice_design_pattern__solid_principle.md)（示例代码工程）中，我们设计了一个 Monitor 监控系统模块，它可以看成是一个简单的 ETL 系统，负责对监控数据进行采集、处理、输出。监控数据来源于在线商场服务集群各个服务，当前通过消息队列模块 Mq 传递到监控系统，经处理后，存储到数据库模块 Db 上。

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2024-04-05-083019.jpg)

我们假设未来要上线一个不支持对接消息队列的服务、结果数据也需要存储到 ClickHouse 以供后续分析，为了应对未来多变的需求，我们有必要将监控系统设计得足够的可扩展。

于是，整个模块被设计为[插件化风格的架构]( https://mp.weixin.qq.com/s/4vofgxx-vasf-957Y_dcww)，`Pipeline` 是数据处理的流水线，其中包含了 `Input`、`Filter` 和 `Output` 三类插件，`Input` 负责从各类数据源中获取监控数据，`Filter` 负责数据处理，`Output` 负责将处理后的数据输出。

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2024-04-05-074731.png)

上述设计中，我们抽象出 `Input`、`Filter` 和 `Output` 三类插件，它们各种往独立的方向变化，最后在 `Pipeline` 上进行灵活组合，这使用桥接模式正合适。

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2024-04-06-010221.png)

##  代码实现

```go
// 关键点1：明确产品的变化点，这里是input、filter和output三类插件，它们各自变化

// demo/monitor/input/input_plugin.go
package input

// 关键点2：将产品的变化点抽象成接口，这里是input.Plugin，filter.Plugin和output.Plugin
// Plugin 输入插件
type Plugin interface {
    plugin.Plugin
    Input() (*plugin.Event, error)
}

// 关键点3：实现产品变化点的接口，这里是SocketInput, AddTimestampFilter和MemoryDbOutput
// demo/monitor/input/socket_input.go
type SocketInput struct {
    socket      network.Socket
    endpoint    network.Endpoint
    packets     chan *network.Packet
    isUninstall uint32
}

func (s *SocketInput) Input() (*plugin.Event, error) {
    packet, ok := <-s.packets
    if !ok {
        return nil, plugin.ErrPluginUninstalled
    }
    event := plugin.NewEvent(packet.Payload())
    event.AddHeader("peer", packet.Src().String())
    return event, nil
}

// demo/monitor/filter/filter_plugin.go
package filter

// Plugin 过滤插件
type Plugin interface {
    plugin.Plugin
    Filter(event *plugin.Event) *plugin.Event
}

// demo/monitor/filter/add_timestamp_filter.go
// AddTimestampFilter 为MonitorRecord增加时间戳
type AddTimestampFilter struct {
}

func (a *AddTimestampFilter) Filter(event *plugin.Event) *plugin.Event {
    re, ok := event.Payload().(*model.MonitorRecord)
    if !ok {
        return event
    }
    re.Timestamp = time.Now().Unix()
    return plugin.NewEvent(re)
}


// demo/monitor/output/output_plugin.go
// Plugin 输出插件
type Plugin interface {
    plugin.Plugin
    Output(event *plugin.Event) error
}

// demo/monitor/output/memory_db_output.go
type MemoryDbOutput struct {
    db        db.Db
    tableName string
}

func (m *MemoryDbOutput) Output(event *plugin.Event) error {
    r, ok := event.Payload().(*model.MonitorRecord)
    if !ok {
    return fmt.Errorf("memory db output unknown event type %T", event.Payload())
    }
    return m.db.Insert(m.tableName, r.Id, r)
}

// 关键点4：定义产品的接口或者实现，通过组合的方式把变化点桥接起来。
// demo/monitor/pipeline/pipeline_plugin.go
// Plugin pipeline由input、filter、output三种插件组成，定义了一个数据处理流程
// 数据流向为 input -> filter -> output
// 如果是接口，可以通过定义Setter方法达到聚合的目的。
type Plugin interface {
    plugin.Plugin
    SetInput(input input.Plugin)
    SetFilter(filter filter.Plugin)
    SetOutput(output output.Plugin)
}

// 如果是结构体，直接把变化点作为成员变量来达到聚合的目的。
type pipelineTemplate struct {
    input   input.Plugin
    filter  filter.Plugin
    output  output.Plugin
    isClose uint32
    run     func()
}

func (p *pipelineTemplate) SetInput(input input.Plugin) {
    p.input = input
}

func (p *pipelineTemplate) SetFilter(filter filter.Plugin) {
    p.filter = filter
}

func (p *pipelineTemplate) SetOutput(output output.Plugin) {
    p.output = output
}

// demo/monitor/pipeline/simple_pipeline.go
// SimplePipeline 简单Pipeline实现，每次运行时新启一个goroutine
type SimplePipeline struct {
    pipelineTemplate
}

```

在本系统中，我们通过配置文件来灵活组合插件，利用反射来实现插件的实例化，实例化的实现使用了[抽象工厂模式](https://mp.weixin.qq.com/s/RqqE6f3N_CzEWjdKluZhHg)，详细的实现方法可参考[【Go实现】实践GoF的23种设计模式：抽象工厂模式](https://mp.weixin.qq.com/s/RqqE6f3N_CzEWjdKluZhHg)。

总结实现桥接模式的几个关键点：

1. 明确产品的变化点，这里是 input、filter 和 output 三类插件，它们各自变化。
2. 将产品的变化点抽象成接口，这里是 `input.Plugin`，`filter.Plugin` 和 `output.Plugin`。
3. 实现产品变化点的接口，这里是 `SocketInput`, `AddTimestampFilter` 和 `MemoryDbOutput`。
4. 定义产品的接口或者实现，通过组合的方式把变化点**桥接**起来。这里是 `pipeline.Plugin` 通过 `Setter` 方法将`input.Plugin`，`filter.Plugin` 和 `output.Plugin` 三个抽象接口桥接了起来。后面即可实现各类 input、filter 和 output 的灵活组合了。

##  扩展

### TiDB 中的桥接模式

[TiDB](https://docs.pingcap.com/zh/tidb/stable/overview) 是一款出色的分布式关系型数据库，它对外提供了一套插件框架，方便用户进行功能扩展。TiDB 的插件框架的设计，也运用到了桥接模式的思想。

![](http://yrunz-1300638001.cos.ap-guangzhou.myqcloud.com/2024-04-13-014221.png)

如上图所示，每个 `Plugin` 都包含 `Validate`、`OnInit`、`OnShutdown`、`OnFlush` 四个待用户实现的接口，它们可以按照各自的方向去变化，然后灵活组合在 `Plugin` 中。

```go
// Plugin presents a TiDB plugin.
type Plugin struct {
    *Manifest
    library  *gplugin.Plugin
    Path     string
    Disabled uint32
    State    State
}

// Manifest describes plugin info and how it can do by plugin itself.
type Manifest struct {
    Name           string
    Description    string
    RequireVersion map[string]uint16
    License        string
    BuildTime      string
    // Validate defines the validate logic for plugin.
    // returns error will stop load plugin process and TiDB startup.
    Validate func(ctx context.Context, manifest *Manifest) error
    // OnInit defines the plugin init logic.
    // it will be called after domain init.
    // return error will stop load plugin process and TiDB startup.
    OnInit func(ctx context.Context, manifest *Manifest) error
    // OnShutDown defines the plugin cleanup logic.
    // return error will write log and continue shutdown.
    OnShutdown func(ctx context.Context, manifest *Manifest) error
    // OnFlush defines flush logic after executed `flush tidb plugins`.
    // it will be called after OnInit.
    // return error will write log and continue watch following flush.
    OnFlush      func(ctx context.Context, manifest *Manifest) error
    flushWatcher *flushWatcher

    Version uint16
    Kind    Kind
}
```

TiDB 在实现插件框架时，使用函数式编程的方式来定义 OnXXX 接口，更具有 Go 风格。

## 典型应用场景

- 从多个维度上对系统/类/结构体进行扩展，如插件化架构。
- 在运行时切换不同的实现，如插件化架构。
- 用于构建与平台无关的程序适配层。

## 优缺点

### 优点

- 可实现抽象不分与实现解耦，变化实现时，客户端无须修改代码，符合[开闭原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA)。
- 每个分离的变化点都可以专注于自身的演进，符合[单一职责原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA)。

### 缺点

- 过度的抽象（过度设计）会使得接口膨胀，导致系统复杂性变大，难以维护。


## 与其他模式的关联

桥接模式通常与[抽象工厂模式](https://mp.weixin.qq.com/s/RqqE6f3N_CzEWjdKluZhHg)搭配使用，比如，在本文例子中，可以通过抽象工厂模式对各个 Plugin 完成实例化，详情见[【Go实现】实践GoF的23种设计模式：抽象工厂模式](https://mp.weixin.qq.com/s/RqqE6f3N_CzEWjdKluZhHg)。

### 文章配图

可以在 [用Keynote画出手绘风格的配图](https://mp.weixin.qq.com/s/-sYW-oa6KzTR9LNdMWCSnQ) 中找到文章的绘图方法。

> #### 参考
>
> [1] [【Go实现】实践GoF的23种设计模式：SOLID原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA), 元闰子
>
> [2] [Design Patterns, Chapter 4. Structural Patterns](https://learning.oreilly.com/library/view/design-patterns-elements/0201633612/ch04.html), GoF
>
> [3] [【Go实现】实践GoF的23种设计模式：抽象工厂模式](https://mp.weixin.qq.com/s/RqqE6f3N_CzEWjdKluZhHg), 元闰子
>
> [4] [桥接模式](https://refactoringguru.cn/design-patterns/bridge), refactoringguru.cn
>
> 更多文章请关注微信公众号：**元闰子的邀请**

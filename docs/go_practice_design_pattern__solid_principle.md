> 之前也有写过关于设计模式的文章《[使用Go实现GoF的23种设计模式](https://mp.weixin.qq.com/mp/appmsgalbum?__biz=Mzg3MjAyNjUyMQ==&action=getalbum&album_id=2280784755992526856#wechat_redirect)》，但是那个系列写了3篇文章就没再继续了，主要的原因是找不到合适的示例代码。考虑到，如果以类似于“鸭子是否会飞”、“烘焙的制作流程”等贴近生活的事情举例，很难在我们日常的开发中产生联系。（目前应该很少有这些逻辑的软件系统吧）
>
> 《**实践GoF的23种设计模式**》可以看成是《使用Go实现GoF的23种设计模式》系列的重启，吸取了上次烂尾的教训，本次在写文章之前就已经完成了23种设计模式的示例代码实现。示例代码**以我们日常开发中经常碰到的一些技术/问题/场景作为切入点**，示范如何运用设计模式来完成相关的实现。

## 前言

从1995年GoF提出23种**设计模式**到现在，25年过去了，设计模式依旧是软件领域的热门话题。设计模式通常被定义为：

> 设计模式（Design Pattern）是一套被反复使用、多数人知晓的、经过分类编目的、代码设计经验的总结，使用设计模式是为了可重用代码、让代码更容易被他人理解并且保证代码可靠性。

从定义上看，**设计模式其实是一种经验的总结，是针对特定问题的简洁而优雅的解决方案**。既然是经验总结，那么学习设计模式最直接的好处就在于可以站在巨人的肩膀上解决软件开发过程中的一些特定问题。

学习设计模式的最高境界是吃透它们本质思想，可以做到**即使已经忘掉某个设计模式的名称和结构，也能在解决特定问题时信手拈来**。设计模式背后的本质思想，就是我们熟知的**SOLID原则**。如果把设计模式类比为武侠世界里的武功招式，那么SOLID原则就是内功内力。通常来说，先把内功练好，再来学习招式，会达到事半功倍的效果。因此，在介绍设计模式之前，很有必要先介绍一下SOLID原则。

本文首先会介绍本系列文章中用到的示例代码demo的整体结构，然后开始逐一介绍SOLID原则，也即单一职责原则、开闭原则、里氏替换原则、接口隔离原则和依赖倒置原则。

## 一个简单的分布式应用系统

> 本系列示例代码demo获取地址：https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation

示例代码demo工程实现了一个简单的分布式应用系统（单机版），该系统主要由以下几个模块组成：

- **网络 Network**，网络功能模块，模拟实现了报文转发、socket通信、http通信等功能。
- **数据库 Db**，数据库功能模块，模拟实现了表、事务、dsl等功能。
- **消息队列 Mq**，消息队列模块，模拟实现了基于topic的生产者/消费者的消息队列。
- **监控系统 Monitor**，监控系统模块，模拟实现了服务日志的收集、分析、存储等功能。
- **边车 Sidecar**，边车模块，模拟对网络报文进行拦截，实现access log上报、消息流控等功能。
- **服务 Service**，运行服务，当前模拟实现了服务注册中心、在线商城服务集群、服务消息中介等服务。

![](https://tva1.sinaimg.cn/large/e6c9d24egy1gzn32jkkduj213g0o00xq.jpg)

示例代码demo工程的主要目录结构如下：

```shell
├── db # 数据库模块，定义Db、Table、TableVisitor等抽象接口和实现
├── monitor # 监控系统模块，采用插件式的架构风格，当前实现access log日志etl功能
│   ├── config  # 监控系统插件配置模块
│   ├── filter # 过滤插件的实现定义
│   ├── input # 输入插件的实现定义
│   ├── output # 输出插件的实现定义
│   ├── pipeline # Pipeline插件的实现定义，一个pipeline表示一个ETL处理流程
│   ├── plugin # 插件抽象接口的定义，比如Plugin、Config等
│   └── model # 监控系统模型对象定义
├── mq # 消息队列模块
├── network  # 网络模块，模拟网络通信，定义了socket、packet等通用类型/接口 
│   └── http # 模拟实现了http通信等服务端、客户端能力
├── service # 服务模块，定义了服务的基本接口
│   ├── mediator # 服务消息中介，作为服务通信的中转方，实现了服务发现，消息转发的能力
│   ├── registry # 服务注册中心，提供服务注册、去注册、更新、 发现、订阅、去订阅、通知等功能
│   │   └── model # 服务注册/发现相关的模型定义
│   └── shopping # 模拟在线商城服务群的定义，包含订单服务、库存服务、支付服务、发货服务
└── sidecar # 边车模块，对socket进行拦截，提供http access log、流控功能
    └── flowctrl # 流控模块，基于消息速率进行随机流控
```

## SRP：单一职责原则

**单一职责原则**（The **S**ingle **R**esponsibility **P**rinciple，**SRP**）应该是SOLID原则中，最容易被理解的一个，但同时也是最容易被误解的一个。很多人会把“将大函数重构成一个个职责单一的小函数”这一重构手法等价为SRP，这是不对的，小函数固然体现了职责单一，但这并不是SRP。

SRP传播最广的定义应该是*Uncle Bob*给出的：

> A module should have one, and only one, reason to change.

也即，**一个模块应该有且只有一个导致其变化的原因**。

这个解释里有2个需要理解的地方：

***（1）如何定义一个模块***

我们通常会把一个源文件定义为最小粒度的模块。

***（2）如何找到这个原因***

一个软件的变化往往是为了满足某个用户的需求，那么这个**用户**就是导致变化的原因。但是，一个模块的用户/客户端程序往往不只一个，比如Java中的ArrayList类，它可能会被成千上万的程序使用，但我们不能说ArrayList职责不单一。因此，我们应该把“一个用户”改为“一类角色”，比如ArrayList的客户端程序都可以归类为“需要链表/数组功能”的角色。

于是，*Uncle Bob*给出了SRP的另一个解释：

> A module should be responsible to one, and only one, actor.

有了这个解释，我们就可以理解**函数职责单一并不等同于SRP**，比如在一个模块有A和B两个函数，它们都是职责单一的，但是函数A的使用者是A类用户，函数B的使用者是B类用户，而且A类用户和B类用户变化的原因都是不一样的，那么这个模块就不满足SRP了。

下面，以我们的分布式应用系统demo为例进一步探讨。对于`Registry`类（服务注册中心）来说，它对外提供的基本能力有服务注册、更新、去注册和发现功能，那么，我们可以这么实现：

```go
// Registry 服务注册中心
type Registry struct {
	db            db.Db
	server        *http.Server
	localIp       string
}

func (r *Registry) register(req *http.Request) *http.Response {...}
func (r *Registry) deregister(req *http.Request) *http.Response {...}
func (r *Registry) update(req *http.Request) *http.Response {...}
func (r *Registry) discovery(req *http.Request) *http.Response {...}

...
```

上述实现中，`Registry`包含了`register`、`update`、`deregister`、`discovery`等4个主要方法，正好对应了`Registry`对外提供的能力，看起来已经是职责单一了。

但是在仔细思考一下就会发现，服务注册、更新和去注册是给专门给**服务提供者**使用的功能，而服务发现则是专门给**服务消费者**使用的功能。**服务提供者和服务消费者是两类不同的角色，它们产生变化的时间和方向都可能不同**。比如：

> 当前服务发现功能是这么实现的：`Registry`从满足查询条件的所有`ServiceProfile`中挑选一个返回给服务消费者（也即`Registry`自己做了负载均衡）。
>
> 假设现在服务消费者提出**新的需求**：`Registry`把所有满足查询条件的`ServiceProfile`都返回，由服务消费者自己来做负载均衡。
>
> 为了实现这样的功能，我们就要修改`Registry`的代码。按理，服务注册、更新、去注册等功能并不应该受到影响，但因为它们和服务发现功能都在同一个模块（`Registry`）里，于是被迫也受到影响了，比如可能会代码冲突。

因此，更好的设计是将`register`、`update`、`deregister`内聚到一个**服务管理**模块`SvcManagement`，`discovery`则放到另一个**服务发现模块**`SvcDiscovery`，服务注册中心`Registry`再组合`svcManagement`和`svcDiscovery`。

![](https://tva1.sinaimg.cn/large/e6c9d24egy1gzqnbif3dcj21he0rcqa0.jpg)

具体实现如下：

```go
// demo/service/registry/svc_management.go
type svcManagement struct {
	db db.Db
	...
}

func (s *svcManagement) register(req *http.Request) *http.Response {...}
func (s *svcManagement) deregister(req *http.Request) *http.Response {...}
func (s *svcManagement) update(req *http.Request) *http.Response {...}


// demo/service/registry/svc_discovery.go
type svcDiscovery struct {
	db db.Db
}

func (s *svcDiscovery) discovery(req *http.Request) *http.Response {...}

// demo/service/registry/registry.go
type Registry struct {
	db            db.Db
	server        *http.Server
	localIp       string
	svcManagement *svcManagement // 组合svcManagement
	svcDiscovery  *svcDiscovery // 组合svcDiscovery
}

func (r *Registry) Run() error {
  ...
  // 调用svcManagement和svcDiscovery完成服务
	return r.server.Put("/api/v1/service-profile", r.svcManagement.register).
		Post("/api/v1/service-profile", r.svcManagement.update).
		Delete("/api/v1/service-profile", r.svcManagement.deregister).
		Get("/api/v1/service-profile", r.svcDiscovery.discovery).
		Put("/api/v1/subscription", r.svcManagement.subscribe).
		Delete("/api/v1/subscription", r.svcManagement.unsubscribe).
		Start()
}

```

除了重复的代码编译，违反SRP还会带来以下2个常见的问题：

1、**代码冲突**。程序员A修改了模块的A功能，而程序员B在不知情的情况下也在修改该模块的B功能（因为A功能和B功能面向不同的用户，完全可能由2位不同的程序员来维护），当他们同时提交修改时，代码冲突就会发生（修改了同一个源文件）。

2、**A功能的修改影响了B功能**。如果A功能和B功能都使用了模块里的一个公共函数C，现在A功能有新的需求需要修改函数C，那么如果修改人没有考虑到B功能，那么B功能的原有逻辑就会受到影响。

由此可见，违反SRP会导致软件的可维护性变得极差。但是，我们也**不能盲目地进行模块拆分，这样会导致代码过于碎片化**，同样也会提升软件的复杂性。比如，在前面的例子中，我们就没有必要再对服务管理模块进行拆分为服务注册模块、服务更新模块和服务去注册模块，一是因为它们面向的用户是一致的；二是在可预见的未来它们要么同时变化，要么都不变。

因此，我们可以得出这样的结论：

1. **如果一个模块面向的都是同一类用户（变化原因一致），那么就没必要进行拆分**。
2. **如果缺乏用户归类的判断，那么最好的拆分时机是变化发生时**。

SRP是聚合和拆分的一个平衡，**太过聚合会导致牵一发动全身，拆分过细又会提升复杂性**。要从用户的视角来把握拆分的度，把面向不同用户的功能拆分开。如果实在无法判断/预测，那就等变化发生时再拆分，避免过度的设计。

## OCP：开闭原则

开闭原则（The **O**pen-**C**lose **P**rinciple，**OCP**）中，“开”指的是对**扩展**开放，“闭”指的是对**修改**封闭，它的完整解释为：

> A software artifact should be open for extension but closed for modification.

通俗地讲就是，**一个软件系统应该具备良好的可扩展性，新增功能应当通过扩展的方式实现，而不是在已有的代码基础上修改**。

然而，从字面意思上看，OCP貌似又是自相矛盾的：想要给一个模块新增功能，但是又不能修改它。

*如何才能打破这个困境呢？*关键是**抽象**！优秀的软件系统总是建立在良好的抽象的基础上，抽象化可以降低软件系统的复杂性。

*那么什么是抽象呢？*抽象不仅存在于软件领域，在我们的生活中也随处可见。下面以《[语言学的邀请](https://book.douban.com/subject/26431646/0)》中的一个例子来解释抽象的含义：

> 假设某农庄有一头叫“阿花”的母牛，那么：
>
> 1、当把它称为“**阿花**”时，我们看到的是它独一无二的一些特征：身上有很多斑点花纹、额头上还有一个闪电形状的伤疤。
>
> 2、当把它称为**母牛**时，我们忽略了它的独有特征，看到的是它与母牛“阿黑”，母牛“阿黄”的共同点：是一头牛、雌性的。
>
> 3、当把它称为**家畜**时，我们又忽略了它作为母牛的特征，而是看到了它和猪、鸡、羊一样的特点：是一个动物，在农庄里圈养。
>
> 4、当把它称为**农庄财产**时，我们只关注了它和农庄上其他可售对象的共同点：可以卖钱、转让。
>
> 从“阿花”，到母牛，到家畜，再到农庄财产，这就是一个不断抽象化的过程。

从上述例子中，我们可以得出这样的结论：

1. **抽象就是不断忽略细节，找到事物间共同点的过程**。
2. **抽象是分层的，抽象层次越高，细节也就越少**。

再回到软件领域，我们也可以把上述的例子类比到**数据库**上，数据库的抽象层次从低至高可以是这样的：`MySQL 8.0版本 -> MySQL -> 关系型数据库 -> 数据库`。现在假设有一个需求，需要业务模块将业务数据保存到数据库上，那么就有以下几种设计方案：

- 方案一：把业务模块设计为直接依赖**MySQL 8.0版本**。因为版本总是经常变化的，如果哪天MySQL升级了版本，那么我们就得修改业务模块进行适配，所以方案一违反了OCP。
- 方案二：把业务模块设计为依赖**MySQL**。相比于方案一，方案二消除了MySQL版本升级带来的影响。现在考虑另一种场景，如果因为某些原因公司禁止使用MySQL，必须切换到PostgreSQL，这时我们还是得修改业务模块进行数据库的切换适配。因此，在这种场景下，方案二也违反了OCP。
- 方案三：把业务模块设计为依赖**关系型数据库**。到了这个方案，我们基本消除了关系型数据库切换的影响，可以随时在MySQL、PostgreSQL、Oracle等关系型数据库上进行切换，而无须修改业务模块。但是，熟悉业务的你预测未来随着用户量的迅速上涨，关系型数据库很有可能无法满足高并发写的业务场景，于是就有了下面的最终方案。
- 方案四：把业务模块设计为依赖**数据库**。这样，不管以后使用MySQL还是PostgreSQL，关系型数据库还是非关系型数据库，业务模块都不需要再改动。到这里，我们基本可以认为业务模块是稳定的，不会受到底层数据库变化带来的影响，满足了OCP。

我们可以发现，上述方案的演进过程，就是我们不断对业务依赖的数据库模块进行抽象的过程，最终设计出稳定的、服务OCP的软件。

那么，在编程语言中，我们用什么来表示“数据库”这一抽象呢？是**接口**！

数据库最常见的几个操作就是CRUD，因此我们可以设计这么一个Db接口来表示“数据库”：

```go
type Db interface {
    Query(tableName string, cond Condition) (*Record, error)
    Insert(tableName string, record *Record) error
    Update(tableName string, record *Record) error
    Delete(tableName string, record *Record) error
}
```

这样，业务模块和数据库模块之间的依赖关系就变成如下图所示：

![](https://tva1.sinaimg.cn/large/e6c9d24egy1gzoyhk01blj216c0o479a.jpg)

满足OCP的另一个关键点就是**分离变化**，只有先把变化点识别分离出来，我们才能对它进行抽象化。下面以我们的分布式应用系统demo为例，解释如何实现变化点的分离和抽象。

在demo中，监控系统主要负责对服务的access log进行ETL操作，也即涉及如下3个操作：1）从消息队列中获取日志数据；2）对数据进行加工；3）将加工后的数据存储在数据库上。

我们把整一个日志数据的处理流程称为pipeline，那么我们可以这么实现：

```go
type Pipeline struct {
    mq Mq
    db Db 
    ...
}

func (p *Pipeline) Run() {
	for atomic.LoadUint32(&p.isClose) != 1 {
    // 1、从消息队列中获取数据
    msg := p.mq.Consume("monitor.topic")
    log := msg.Payload()
    
    // 2、对数据进行字段提取操作
    matches := p.pattern.FindStringSubmatch(log)
	  if len(matches) != 3 {
		  return event
	  }
	  record := model.NewMonitoryRecord()
	  record.Endpoint = matches[1]
	  record.Type = model.Type(matches[2])
    
    // 3.存储到数据库上
    p.db.Insert("logs_table", record.Id, record)
    ...
	}
}
```

现在考虑新上线一个服务，但是这个服务不支持对接消息队列了，只支持socket传输数据，于是我们得在`Pipeline`上新增一个`InputType`来判断是否使用socket输入源：

```go
func (p *Pipeline) Run() {
	for atomic.LoadUint32(&p.isClose) != 1 {
    if inputType == input.MqType {
        // 从消息队列中获取数据
        msg := p.mq.Consume("monitor.topic")
        log := msg.Payload()
    } else {
      // 使用socket为消息来源
      packet := socket.Receice()
      log := packet.PayLoad().(string)
    }
    ...
	}
}
```

过一段时间，有需求需要给access log打上一个时间戳，方便后续的日志分析，于是我们需要修改`Pipeline`的数据加工逻辑：

```go
func (p *Pipeline) Run() {
	for atomic.LoadUint32(&p.isClose) != 1 {
    ...
    // 对数据进行字段提取操作
    matches := p.pattern.FindStringSubmatch(log)
	  if len(matches) != 3 {
		  return event
	  }
	  record := model.NewMonitoryRecord()
	  record.Endpoint = matches[1]
	  record.Type = model.Type(matches[2])
    // 新增一个时间戳字段
    record.Timestamp = time.Now().Unix()
    ...
	}
}
```

很快，又有一个需求，需要将加工后的数据存储到ES上，方便后续的日志检索，于是我们再次修改了`Pipeline`的数据存储逻辑：

```go
func (p *Pipeline) Run() {
	for atomic.LoadUint32(&p.isClose) != 1 {
    ...
    if outputType == output.Db {
        // 存储到ES上
        p.db.Insert("logs_table", record.Id, record)
    } else {
        // 存储到ES上
        p.es.Store(record.Id, record)
    }
    ...
	}
}
```

在上述的pipeline例子中，每次新增需求都需要修改`Pipeline`模块，明显违反了OCP。下面，我们来对它进行优化，使它满足OCP。

  第一步是**分离变化点**，根据pipeline的业务处理逻辑，我们可以发现3个独立的变化点，数据的获取、加工和存储。第二步，我们对这3个变化点进行抽象，设计出以下3个抽象接口：

```go
// demo/monitor/input/input_plugin.go
// 数据获取抽象接口
type InputPlugin interface {
	plugin.Plugin
	Input() (*plugin.Event, error)
}

// demo/monitor/filter/filter_plugin.go
// 数据加工抽象接口
type FilterPlugin interface {
	plugin.Plugin
	Filter(event *plugin.Event) *plugin.Event
}

// demo/monitor/output/output_plugin.go
// 数据存储抽象接口
type OutputPlugin interface {
	plugin.Plugin
	Output(event *plugin.Event) error
}
```

最后，`Pipeline`的实现如下，只依赖于`InputPlugin`、`FilterPlugin`和`OutputPlugin`三个抽象接口。后续再有需求变更，只需扩展对应的接口即可，`Pipeline`无须再变更：

```go
// demo/monitor/pipeline/pipeline_plugin.go
// ETL流程定义
type pipelineTemplate struct {
	input   input.Plugin
	filter  filter.Plugin
	output  output.Plugin
  ...
}

func (p *pipelineTemplate) doRun() {
  ...
	for atomic.LoadUint32(&p.isClose) != 1 {
		event, err := p.input.Input()
		event = p.filter.Filter(event)
		p.output.Output(event)
	}
  ...
}

```

![](https://tva1.sinaimg.cn/large/e6c9d24egy1gzpffjrzg7j21h00s4ahv.jpg)

**OCP是软件设计的终极目标**，我们都希望能设计出可以新增功能却不用动老代码的软件。但是100%的对修改封闭肯定是做不到的，另外，遵循OCP的代价也是巨大的。它需要软件设计人员能够根据具体的业务场景识别出那些最有可能变化的点，然后分离出去，抽象成稳定的接口。这要求设计人员必须具备丰富的实战经验，以及非常熟悉该领域的业务场景。否则，**盲目地分离变化点、过度地抽象，都会导致软件系统变得更加复杂**。

## LSP：里氏替换原则

上一节介绍中，OCP的一个关键点就是**抽象**，而如何判断一个抽象是否合理，这是**里氏替换原则**（The **L**iskov **S**ubstitution **P**rinciple，**LSP**）需要回答的问题。

LSP的最初定义如下：

> If for each object o1 of type S there is an object o2 of type T such that for all programs P defined in terms of T, the behavior of P is unchanged when o1 is substituted for o2 then S is a subtype of T.

简单地讲就是，**子类型必须能够替换掉它们的基类型**，也即基类中的所有性质，在子类中仍能成立。一个简单的例子：假设有一个函数f，它的入参类型是基类B。同时，基类B有一个派生类D，如果把D的实例传递给函数f，那么函数f的行为功能应该是不变的。

由此可以看出，违反LSP的后果很严重，会导致程序出现不在预期之内的行为错误。最典型的就是正方形继承自长方形的例子。

> 长方形和正方形例子的详细介绍，请参考[【Java实现】实践GoF的23种设计模式：SOLID原则](https://mp.weixin.qq.com/s/0vIrrjgPGUixDnUe6nQ5Cw) 中的“LSP：里氏替换原则”一节

出现违反LSP的设计，主要原因还是我们**孤立地进行模型设计**，没有从客户端程序的角度来审视该设计是否正确。我们孤立地认为在数学上成立的关系（正方形 IS-A 矩形），在程序中也一定成立，而忽略了客户端程序的使用方法。

该例子告诉我们：**一个模型的正确性或有效性，只能通过客户端程序来体现**。

下面，我们总结一下在继承体系（IS-A）下，要想设计出符合LSP的模型所需要遵循的一些约束：

1. **基类应该设计为一个抽象类**（不能直接实例化，只能被继承）。
2. **子类应该实现基类的抽象接口，而不是重写基类已经实现的具体方法**。
3. **子类可以新增功能，但不能改变基类的功能**。
4. **子类不能新增约束**，包括抛出基类没有声明的异常。

前面的矩形和正方形的例子中，几乎把这些约束都打破了，从而导致了程序的异常行为：1）`Square`的基类`Rectangle`不是一个抽象类，打破`约束1`；2）`Square`重写了基类的`setWidth`和`setLength`方法，打破`约束2`；3）`Square`新增了`Rectangle`没有的约束，长宽相等，打破`约束4`。

除了继承之外，另一个实现抽象的机制是**接口**。如果我们是面向接口的设计，那么上述的`约束1～3`其实已经满足了：1）接口本身不具备实例化能力，满足`约束1`；2）接口没有具体的实现方法（*Java中接口的default方法比较例外，本文先不考虑*），也就不会被重写，满足`约束2`；3）接口本身只定义了行为契约，并没有实际的功能，因此也不会被改变，满足`约束3`。

因此，使用接口替代继承来实现多态和抽象，能够减少很多不经意的错误。但是面向接口设计仍然需要遵循`约束4`，下面我们以分布式应用系统demo为例，介绍一个比较隐晦地打破`约束4`，从而违反了LSP的实现。

还是以监控系统为例，为例实现ETL流程的灵活配置，我们需要通过配置文件定义pipeline的流程功能（数据从哪获取、需要经过哪些加工、加工后存储到哪里）。当前需要支持json和yaml两种配置文件格式，以yaml配置为例，配置内容是这样的：

```yaml
# src/main/resources/pipelines/pipeline_0.yaml
name: pipeline_0 # pipeline名称
type: single_thread # pipeline类型
input: # input插件定义（数据从哪里来）
  name: input_0 # input插件名称
  type: memory_mq # input插件类型
  context: # input插件的初始化上下文
    topic: access_log.topic
filter: # filter插件定义（需要经过哪些加工）
  - name: filter_0 # 加工流程filter_0定义，类型为log_to_json
    type: log_to_json
  - name: filter_1 # 加工流程filter_1定义，类型为add_timestamp
    type: add_timestamp
  - name: filter_2 # 加工流程filter_2定义，类型为json_to_monitor_event
    type: json_to_monitor_event
output: # output插件定义（加工后存储到哪里）
  name: output_0 # output插件名称
  type: memory_db # output插件类型
  context: # output插件的初始化上下文
    tableName: monitor_event_0
```

首先我们定义一个`Config`接口来表示“配置”这一抽象:

```go
// demo/monitor/plugin/plugin.go
packet plugin
// Config 插件配置抽象接口
type Config interface {
	Load(conf string) error
}
```

另外，上述配置中的`input`、`filter`、`output`子项，可以认为是`input.Plugin`、`filter.Plugin`、`output.Plugin`插件的配置项，由`Pipeline`插件的配置项组合在一起，因此我们定义了如下几个`Config`的实现类：

```go
// demo/monitor/config/config.go
package config

type item struct {
	Name       string         `json:"name" yaml:"name"`
	PluginType string         `json:"type" yaml:"type"`
	Ctx        plugin.Context `json:"context" yaml:"context"`
	loadConf   func(conf string, item interface{}) error // 区分yaml和json的加载方式
}

// 输入插件配置
type Input item
func (i *Input) Load(conf string) error {
	return i.loadConf(conf, i)
}

// 过滤插件配置
type Filter item
func (f *Filter) Load(conf string) error {
	return f.loadConf(conf, f)
}

// 输出插件配置
type Output item
func (o *Output) Load(conf string) error {
	return o.loadConf(conf, o)
}

// Pipeline插件配置
type Pipeline struct {
	item    `yaml:",inline"` // yaml嵌套时需要加上,inline
	Input   Input            `json:"input" yaml:"input"`
	Filters []Filter         `json:"filters" yaml:"filters,flow"`
	Output  Output           `json:"output" yaml:"output"`
}
func (p *Pipeline) Load(conf string) error {
	return p.loadConf(conf, p)
}
```

因为涉及到从配置到对象的实例化过程，自然会想到使用***工厂方法模式***来创建对象。另外因为`pipeline.Pipeline`、`input.Plugin`、`filter.Plugin`和`output.Plugin`都实现了`Plugin`接口，我们也很容易想到定义一个`PluginFactory`接口来表示“插件工厂”这一抽象，具体的插件工厂再实现该接口：

```go
// 插件工厂接口，根据配置实例化插件
type PluginFactory interface {
    Create(config plugin.Config) plugin.Plugin
}

// input插件工厂
type InputPluginFactory struct {}
func (i *InputPluginFactory) Create(config plugin.Config) plugin.Plugin {
    conf := config.(*config.Input)
    ... // input插件实例化过程
}
// filter插件工厂
type FilterPluginFactory struct {}
func (f *FilterPluginFactory) Create(config plugin.Config) plugin.Plugin {
    conf := config.(*config.Filter)
    ... // filter插件实例化过程
}

// output插件工厂
type OutputPluginFactory struct {}
func (o *OutputPluginFactory) Create(config plugin.Config) plugin.Plugin {
    conf := config.(*config.Output)
    ... // output插件实例化过程
}

// pipeline插件工厂
type PipelinePluginFactory struct {}
func (p *PipelinePluginFactory) Create(config plugin.Config) plugin.Plugin {
    conf := config.(*config.Pipeline)
    ... // pipeline插件实例化过程
}


```

最后，通过`PipelineFactory`来创建`Pipline`对象：

```go
conf := &config.Pipeline{...}
pipeline := NewPipelinePluginFactory().Create(conf)
...
```

到目前为止，上述的设计看起来是合理的，运行也没有问题。

但是，细心的读者可能会发现，每个插件工厂子类的`Create`方法的第一行代码都是一个转型语句，比如`PipelineFactory`的是`conf := config.(*config.Pipeline)`。所以，上一段代码能够正常运行的前提是：传入`PipelineFactory.Create`方法的入参必须是`*config.Pipeline` 。如果客户端程序传入`*config.Input`的实例，`PipelineFactory.Create`方法将会抛出转型失败的异常。

上述这个例子就是一个违反LSP的典型场景，虽然在约定好的前提下，程序可以运行正确，但是如果有客户端不小心破坏了这个约定，就会带来程序行为异常（我们永远无法预知客户端的所有行为）。

要纠正这个问题也很简单，就是去掉`PluginFactory`这一层抽象，让`PipelineFactory.Create`等工厂方法的入参声明为具体的配置类，比如`PipelineFactory`可以这么实现：

```go
// pipeline插件工厂
type PipelinePluginFactory struct {}
func (p *PipelinePluginFactory) Create(config *config.Pipeline) plugin.Plugin {
    ... // pipeline插件实例化过程
}
```

![](https://tva1.sinaimg.cn/large/e6c9d24egy1gzpopcctvzj214u0mc436.jpg)

从上述几个例子中，我们可以看出遵循LSP的重要性，而设计出符合LSP的软件的要点就是，**根据该软件的使用者行为作出的合理假设，以此来审视它是否具备有效性和正确性**。

## ISP：接口隔离原则

**接口隔离原则**（The **I**nterface **S**egregation **P**rinciple，**ISP**）是关于接口设计的一项原则，这里的“接口”并不单指Java或Go上使用interface声明的狭义接口，而是包含了狭义接口、抽象类、具象类等在内的广义接口。它的定义如下：

> Client should not be forced to depend on methods it does not use.

也即，**一个模块不应该强迫客户程序依赖它们不想使用的接口**，模块间的关系应该建立在最小的接口集上。

下面，我们通过一个例子来详细介绍ISP。

![](https://tva1.sinaimg.cn/large/e6c9d24egy1gzq2gxj891j213a0mowhz.jpg)

上图中，`Client1`、`Client2`、`Client3`都依赖了`Class1`，但实际上，`Client1`只需使用`Class1.func1`方法，`Client2`只需使用`Class1.func2`，`Client3`只需使用`Class1.func3`，那么这时候我们就可以说该设计违反了ISP。

违反ISP主要会带来如下2个问题：

1. **增加模块与客户端程序的依赖**，比如在上述例子中，虽然`Client2`和`Client3`都没有调用`func1`，但是当`Class1`修改`func1`还是必须通知`Client1～3`，因为`Class1`并不知道它们是否使用了`func1`。
2. **产生接口污染**，假设开发`Client1`的程序员，在写代码时不小心把`func1`打成了`func2`，那么就会带来`Client1`的行为异常。也即`Client1`被`func2`给污染了。

为了解决上述2个问题，我们可以把`func1`、`func2`、`func3`通过接口隔离开：

![](https://tva1.sinaimg.cn/large/e6c9d24egy1gzq2zy32hkj212e0okwim.jpg)

接口隔离之后，`Client1`只依赖了`Interface1`，而`Interface1`上只有`func1`一个方法，也即`Client1`不会受到`func2`和`func3`的污染；另外，当`Class1`修改`func1`之后，它只需通知依赖了`Interface1`的客户端即可，大大降低了模块间耦合。

实现ISP的关键是**将大接口拆分成小接口**，而拆分的关键就是**接口粒度的把握**。想要拆分得好，就要求接口设计人员对业务场景非常熟悉，对接口使用的场景了如指掌。否则孤立地设计接口，很难满足ISP。

下面，我们以分布式应用系统demo为例，来进一步介绍ISP的实现。

一个消息队列模块通常包含生产（produce）和消费（consumer）两种行为，因此我们设计了`Mq`消息队列抽象接口，包含`Produce`和`Consume`两个方法：

```java
// Mq 消息队列接口
type Mq interface {
	Consume(topic Topic) (*Message, error)
	Produce(message *Message) error
}

// 当前提供MemoryMq内存消息队列的实现
type MemoryMq struct {...}
func (m *memoryMq) Consume(topic Topic) (*Message, error) {...}
func (m *memoryMq) Produce(message *Message) error {...}
```

当前demo中使用接口的模块有2个，分别是作为消费者的`MemoryMqInput`和作为生产者的`AccessLogSidecar`：

```go
type MemoryMqInput struct {
	topic    mq.Topic
	consumer mq.Mq // 同时依赖了Consume和Produce
}

type AccessLogSidecar struct {
	socket   network.Socket
	producer mq.Mq // 同时依赖了Consume和Produce
	topic    mq.Topic
}
```

**从领域模型上看**，`Mq`接口的设计确实没有问题，它就应该包含`consume`和`produce`两个方法。但是**从客户端程序的角度上看**，它却违反了ISP，对`MemoryMqInput`来说，它只需要`consume`方法；对`AccessLogSidecar`来说，它只需要`produce`方法。

一种设计方案是把`Mq`接口拆分成2个子接口`Consumable`和`Producible`，让`MemoryMq`直接实现`Consumable`和`Producible`：

```go
// Consumable 消费接口，从消息队列中消费数据
type Consumable interface {
	Consume(topic Topic) (*Message, error)
}

// Producible 生产接口，向消息队列生产消费数据
type Producible interface {
	Produce(message *Message) error
}
```

仔细思考一下，就会发现上面的设计不太符合消息队列的领域模型，因为`Mq`的这个抽象确实应该存在的。

更好的设计应该是保留`Mq`抽象接口，让`Mq`继承自`Consumable`和`Producible`，这样的**分层设计**之后，既能满足ISP，又能让实现符合消息队列的领域模型：

![](https://tva1.sinaimg.cn/large/e6c9d24egy1gzq5s8k4rwj21da0pogrf.jpg)

具体实现如下：

```go
// Mq 消息队列接口，继承了Consumable和Producible，同时又consume和produce两种行为
type Mq interface {
	Consumable
	Producible
}

type MemoryMqInput struct {
	topic    mq.Topic
	consumer mq.Consumable // 只依赖Consumable
}

type AccessLogSidecar struct {
	socket   network.Socket
	producer mq.Producible // 只依赖Producible
	topic    mq.Topic
}
```

接口隔离可以减少模块间耦合，提升系统稳定性，但是**过度地细化和拆分接口，也会导致系统的接口数量的上涨，从而产生更大的维护成本**。接口的粒度需要根据具体的业务场景来定，可以参考单一职责原则，**将那些为同一类客户端程序提供服务的接口合并在一起**。

## DIP：依赖倒置原则

《Clean Architecture》中介绍OCP时有提过：**如果要模块A免于模块B变化的影响，那么就要模块B依赖于模块A**。这句话貌似是矛盾的，模块A需要使用模块B的功能，怎么会让模块B反过来依赖模块A呢？这就是**依赖倒置原则**（The **D**ependency **I**nversion **P**rinciple，**DIP**）所要解答的问题。

DIP的定义如下：

> 1. High-level modules should not import anything from low-level modules. Both should depend on abstractions.
> 2. Abstractions should not depend on details. Details (concrete implementations) should depend on abstractions.

翻译过来，就是：

> 1、高层模块不应该依赖低层模块，两者都应该依赖抽象
>
> 2、抽象不应该依赖细节，细节应该依赖抽象

在DIP的定义里，出现了**高层模块**、**低层模块**、**抽象**、**细节**等4个关键字，要弄清楚DIP的含义，理解者4个关键字至关重要。

***（1）高层模块和低层模块***

一般地，我们认为**高层模块是包含了应用程序核心业务逻辑、策略的模块**，是整个应用程序的灵魂所在；**低层模块通常是一些基础设施**，比如数据库、Web框架等，它们主要为了辅助高层模块完成业务而存在。

***（2）抽象和细节***

在前文“OCP：开闭原则”一节中，我们可以知道，**抽象就是众多细节中的共同点**，抽象就是不断忽略细节的出来的。

现在再来看DIP的定义，对于第2点我们不难理解，从抽象的定义来看，抽象是不会依赖细节的，否则那就不是抽象了；而细节依赖抽象往往都是成立的。

**理解DIP的关键在于第1点**，按照我们正向的思维，高层模块要借助低层模块来完成业务，这必然会导致高层模块依赖低层模块。但是在软件领域里，我们可以把这个依赖关系**倒置**过来，这其中的关键就是**抽象**。我们可以忽略掉低层模块的细节，抽象出一个稳定的接口，然后让高层模块依赖该接口，同时让低层模块实现该接口，从而实现了依赖关系的倒置：

![](https://tva1.sinaimg.cn/large/e6c9d24egy1gzq8h7t8i4j21d20qa7a2.jpg)

之所以要把高层模块和底层模块的依赖关系倒置过来，主要是因为作为核心的高层模块不应该受到低层模块变化的影响。**高层模块的变化原因应当只能有一个，那就是来自软件用户的业务变更需求**。

下面，我们通过分布式应用系统demo来介绍DIP的实现。

对于服务注册中心`Registry`来说，当有新的服务注册上来时，它需要把服务信息（如服务ID、服务类型等）保存下来，以便在后续的服务发现中能够返回给客户端。因此，`Registry`需要一个数据库来辅助它完成业务。刚好，我们的数据库模块实现了一个内存数据库`MemoryDb`，于是我们可以这么实现`Registry`：

```go
type Registry struct {
	db            db.MemoryDb // 直接依赖MemoryDb
	server        *http.Server
	localIp       string
	svcManagement *svcManagement
	svcDiscovery  *svcDiscovery
}
```

按照上面的设计，模块间的依赖关系是`Registry`依赖于`MemoryDb`，也即高层模块依赖于低层模块。这种依赖关系是脆弱的，如果哪天需要把存储服务信息的数据库从`MemoryDb`改成`DiskDb`，那么我们也得改`Registry`的代码：

```go
type Registry struct {
	db            db.DiskDb // 改成依赖DiskDb
	server        *http.Server
	localIp       string
	svcManagement *svcManagement
	svcDiscovery  *svcDiscovery
}
```

更好的设计应该是把`Registry`和`MemoryDb`的依赖关系倒置过来，首先我们需要从细节`MemoryDb`抽象出一个稳定的接口`Db`：

更好的设计应该是把`Registry`和`MemoryDb`的依赖关系倒置过来，首先我们需要从细节`MemoryDb`抽象出一个稳定的接口`Db`：

```go
// Db 数据库抽象接口
type Db interface {
	Query(tableName string, primaryKey interface{}, result interface{}) error
	Insert(tableName string, primaryKey interface{}, record interface{}) error
	Update(tableName string, primaryKey interface{}, record interface{}) error
	Delete(tableName string, primaryKey interface{}) error
	...
}
```

接着，我们让`Registry`依赖`Db`接口，而`MemoryDb`实现`Db`接口，以此来完成依赖倒置：

```go
type Registry struct {
	db            db.Db // 依赖Db抽象接口
	server        *http.Server
	localIp       string
	svcManagement *svcManagement
	svcDiscovery  *svcDiscovery
}

// MemoryDb 实现Db接口
type MemoryDb struct {
	tables sync.Map // key为tableName，value为table
}
func (m *memoryDb) Query(tableName string, primaryKey interface{}, result interface{}) error {...}
func (m *memoryDb) Insert(tableName string, primaryKey interface{}, record interface{}) error {...}
func (m *memoryDb) Update(tableName string, primaryKey interface{}, record interface{}) error {...}
func (m *memoryDb) Delete(tableName string, primaryKey interface{}) error {...}
```

当高层模块依赖抽象接口时，总得在某个时候，某个地方把实现细节（低层模块）**注入**到高层模块上。在上述例子中，我们选择在main函数上，在创建`Registry`对象时，把`MemoryDb`注入进去。

![](https://tva1.sinaimg.cn/large/e6c9d24egy1gzqmzskc5nj21fs0qgjxw.jpg)

一般地，我们都会在main/启动函数上完成依赖注入，常见的注入的方式有以下几种：

- 构造函数注入（`Registry`所使用的方法）
- setter方法注入
- 提供依赖注入的接口，客户端直调用该接口即可
- 通过框架进行注入，比如Spring框架中的注解注入能力

另外，DIP不仅仅适用于模块/类/接口设计，在架构层面也同样适用，比如DDD的分层架构和Uncle Bob的整洁架构，都是运用了DIP：

![](https://tva1.sinaimg.cn/large/e6c9d24egy1gzqap4ml4fj21by0rqqac.jpg)

当然，DIP并不是说高层模块是只能依赖抽象接口，它的本意应该是依赖**稳定**的接口/抽象类/具象类。如果一个具象类是稳定的，比如Java中的`String`，那么高层模块依赖它也没有问题；相反，如果一个抽象接口是不稳定的，经常变化，那么高层模块依赖该接口也是违反DIP的，这时候应该思考下接口是否抽象合理。

## 最后

本文花了很长的篇幅讨论了23种设计模式背后的核心思想 —— **SOLID原则**，它能指导我们设计出高内聚、低耦合的软件系统。但是它毕竟只是**原则**，如何落地到实际的工程项目上，还是需要参考成功的实践经验。而这些实践经验正是接下来我们要探讨的**设计模式**。

学习设计模式最好的方法就是**实践**，在《实践GoF的23种设计模式》后续的文章，我们将以本文介绍的[分布式应用系统demo](https://github.com/ruanrunxue/Practice-Design-Pattern--Java-Implementation)作为实践示范，介绍23种设计模式的程序结构、适用场景、实现方法、优缺点等，让大家对设计模式有个更深入的理解，能够**用对**、**不滥用**设计模式。

> **参考**
>
> 1. [Clean Architecture](https://book.douban.com/subject/26915970/), Robert C. Martin (“Uncle Bob”) 
> 2. [敏捷软件开发：原则、模式与实践](https://book.douban.com/subject/1140457/), Robert C. Martin (“Uncle Bob”) 
> 3. [使用Go实现GoF的23种设计模式](https://mp.weixin.qq.com/mp/appmsgalbum?__biz=Mzg3MjAyNjUyMQ==&action=getalbum&album_id=2280784755992526856#wechat_redirect), 元闰子
> 3. [【Java实现】实践GoF的23种设计模式：SOLID原则](https://mp.weixin.qq.com/s/0vIrrjgPGUixDnUe6nQ5Cw) , 元闰子
> 4. [SOLID原则精解之里氏替换原则LSP](https://bbs.huaweicloud.com/blogs/detail/178619), 人民副首席码仔
>
> 更多文章请关注微信公众号：**元闰子的邀请**

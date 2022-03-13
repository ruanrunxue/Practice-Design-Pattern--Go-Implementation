# 实践GoF的23种设计模式：Go实现

## 文章目录

| 归类       |                 模式                  |                           示例代码                           | 文章 |
| ---------- | :-----------------------------------: | :----------------------------------------------------------: | :--: |
| SOLID原则  |          单一职责原则（SRP）          | [Registry](demo/src/main/java/com/yrunz/designpattern/service/registry/Registry.java) |      |
|            |            开闭原则（OCP）            | [Pipeline](demo/src/main/java/com/yrunz/designpattern/monitor/pipeline/Pipeline.java) |      |
|            |          里氏替换原则（LSP）          | [PipelineFactory](demo/src/main/java/com/yrunz/designpattern/monitor/pipeline/PipelineFactory.java) |      |
|            |          接口隔离原则（ISP）          | [Mq](demo/src/main/java/com/yrunz/designpattern/mq/MemoryMq.java) |      |
|            |          依赖倒置原则（DIP）          | [Db](demo/src/main/java/com/yrunz/designpattern/db/Db.java)  |      |
| 创建型模式 |         单例模式（Singleton）         | [Network](demo/src/main/java/com/yrunz/designpattern/network/Network.java) |      |
|            |         建造者模式（Builder）         | [ServiceProfile.Builder](demo/src/main/java/com/yrunz/designpattern/service/registry/model/ServiceProfile.java) |      |
|            |    工厂方法模式（Factory Method）     | [SidecarFactory](demo/src/main/java/com/yrunz/designpattern/sidecar/SidecarFactory.java) |      |
|            |   抽象工厂模式（Abstract Factory）    | [ConfigFactory](demo/src/main/java/com/yrunz/designpattern/monitor/config/ConfigFactory.java) |      |
|            |         原型模式（Prototype）         | [Cloneable](demo/src/main/java/com/yrunz/designpattern/service/registry/model/Cloneable.java) |      |
| 结构型模式 |         适配器模式（Adapter）         | [DslResultRender](demo/src/main/java/com/yrunz/designpattern/db/console/DslResultRender.java) |      |
|            |          桥接模式（Bridge）           | [Pipeline](demo/src/main/java/com/yrunz/designpattern/monitor/pipeline/Pipeline.java) |      |
|            |         组合模式（Composite）         | [Pipeline](demo/src/main/java/com/yrunz/designpattern/monitor/pipeline/Pipeline.java) |      |
|            |        装饰者模式（Decorator）        | [FlowCtrlSidecar](demo/src/main/java/com/yrunz/designpattern/sidecar/FlowCtrlSidecar.java) |      |
|            |          外观模式（Facade）           | [ShoppingCenter](demo/src/main/java/com/yrunz/designpattern/service/shopping/ShoppingCenter.java) |      |
|            |         享元模式（Flyweight）         | [RegionTable](demo/src/main/java/com/yrunz/designpattern/service/registry/model/schema/RegionTable.java) |      |
|            |           代理模式（Proxy）           | [CacheDbProxy](demo/src/main/java/com/yrunz/designpattern/db/cache/CacheDbProxy.java) |      |
| 行为模式   | 责任链模式（Chain Of Responsibility） | [FilterChain](demo/src/main/java/com/yrunz/designpattern/monitor/filter/FilterChain.java) |      |
|            |          命令模式（Command）          | [Command](demo/src/main/java/com/yrunz/designpattern/db/transaction/Command.java) |      |
|            |        迭代器模式（Iterator）         | [TableIterator](demo/src/main/java/com/yrunz/designpattern/db/TableIterator.java) |      |
|            |        中介者模式（Mediator）         | [Mediator](demo/src/main/java/com/yrunz/designpattern/service/mediator/Mediator.java) |      |
|            |         备忘录模式（Memento）         | [CmdHistory](demo/src/main/java/com/yrunz/designpattern/db/transaction/CmdHistory.java) |      |
|            |        观察者模式（Observer）         | [SocketImpl](demo/src/main/java/com/yrunz/designpattern/network/SocketImpl.java) |      |
|            |           状态模式（State）           | [FcState](demo/src/main/java/com/yrunz/designpattern/sidecar/flowctrl/FcState.java) |      |
|            |         策略模式（Strategy）          | [InputPlugin](demo/src/main/java/com/yrunz/designpattern/monitor/input/InputPlugin.java) |      |
|            |    模板方法模式（Template Method）    | [AbstractFcState](demo/src/main/java/com/yrunz/designpattern/sidecar/flowctrl/AbstractFcState.java) |      |
|            |         访问者模式（Visitor）         | [TableVisitor](demo/src/main/java/com/yrunz/designpattern/db/TableVisitor.java) |      |

## 示例代码demo介绍

示例代码demo工程实现了一个简单的分布式应用系统（单机版），该系统主要由以下几个模块组成：

- **网络 Network**，网络功能模块，模拟实现了报文转发、socket通信、http通信等功能。
- **数据库 Db**，数据库功能模块，模拟实现了表、事务、dsl等功能。
- **消息队列 Mq**，消息队列模块，模拟实现了基于topic的生产者/消费者的消息队列。
- **监控系统 Monitor**，监控系统模块，模拟实现了服务日志的收集、分析、存储等功能。
- **边车 Sidecar**，边车模块，模拟对网络报文进行拦截，实现access log上报、消息流控等功能。
- **服务 Service**，运行服务，当前模拟实现了服务注册中心、在线商城服务集群、服务消息中介等服务。

![](https://tva1.sinaimg.cn/large/e6c9d24egy1gzn32jkkduj213g0o00xq.jpg)

主要目录结构如下：

```shell
├── db # 数据库模块，定义Db、Table、TableVisitor等抽象接口和实现
├── monitor # 监控系统模块，采用插件式的架构风格，当前实现access log日志etl功能
│   ├── config  # 监控系统插件配置模块
│   ├── filter # 过滤插件的实现定义
│   ├── input # 输入插件的实现定义
│   ├── output # 输出插件的实现定义
│   ├── pipeline # Pipeline插件的实现定义，一个pipeline表示一个ETL处理流程
│   ├── plugin # 插件抽象接口的定义，比如Plugin、Config等
│   └── record
├── mq
├── network
│   └── http
├── service
│   ├── mediator
│   ├── registry
│   │   └── model
│   └── shopping
└── sidecar
    └── flowctrl
```


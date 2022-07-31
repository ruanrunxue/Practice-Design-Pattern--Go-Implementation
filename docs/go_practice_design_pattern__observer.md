> 上一篇：[【Go实现】实践GoF的23种设计模式：装饰者模式](https://mp.weixin.qq.com/s/NT6_KOY_hGkA-y2b4fw45A)
>
> **简单的分布式应用系统**（示例代码工程）：[https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation)

## 简介

现在有 2 个服务，Service A 和 Service B，通过 REST 接口通信；Service A 在某个业务场景下调用 Service B 的接口完成一个计算密集型任务，假设接口为 http://service_b/api/v1/domain；该任务运行时间很长，但 Service A 不想一直阻塞在接口调用上。为了满足 Service A 的要求，通常有 2 种方案：

1. Service A 隔一段时间调用一次 Service B 的接口，如果任务还没完成，就返回 HTTP Status 102 Processing；如果已完成，则返回 HTTP Status 200 Ok。

   ![](https://tva1.sinaimg.cn/large/e6c9d24egy1h4gk1oezcrj21co0rwtfr.jpg)

2. Service A 在请求 Service B 接口时带上 callback uri，比如 http://service_b/api/v1/domain?callbackuri=http://service_a/api/v1/domain，Service B 收到请求后立即返回 HTTP Status 200 Ok，等任务完成后再调用 Service A callback uri 进行通知。

   ![](https://tva1.sinaimg.cn/large/e6c9d24egy1h4gk5xc6jcj21ae0ryjxj.jpg)

方案 1 须要轮询接口，轮询太频繁会导致资源浪费，间隔太长又会导致任务完成后 Service A 无法及时感知。显然，方案 2 更加高效，因此也被广泛应用。

方案 2 用到的思想就是本文要介绍的**观察者模式**（**Observer Pattern**），GoF 对它的定义如下：

> Define a one-to-many dependency between objects so that when one object changes state, all its dependents are notified and updated automatically.

我们将观察者称为 **Observer**，被观察者（或主体）称为 **Subject**，那么 **Subject 和 Observer 是一对多的关系**，当 Subject 状态变更时，所有的 Observer 都会被通知到。

## UML 结构

![](https://tva1.sinaimg.cn/large/e6c9d24egy1h4gn9y12alj21fm0q2wkz.jpg)

## 场景上下文

在 [简单的分布式应用系统](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation/blob/main/docs/go_ractice_design_pattern__solid_principle.md)（示例代码工程）中，应用之间通过 network 模块来通信，其中通信模型采用观察者模式：

![](https://tva1.sinaimg.cn/large/e6c9d24egy1h4gqq5hw9tj21ea0p2grn.jpg)

从上图可知，App 直接依赖 http 模块，而 http 模块底层则依赖 socket 模块：

1. 在 `App2` 初始化时，先向 http 模块注册一个 `request handler`，处理 `App1` 发送的 http 请求。
2. http 模块会将 `request handler` 转换为 `packet handler` 注册到 socket 模块上。
3. `App 1` 发送 http 请求，http 模块将请求转换为 `socket packet` 发往 `App 2` 的 socket 模块。
4. `App 2` 的 socket 模块收到 packet 后，调用 `packet handler` 处理该报文；`packet handler` 又会调用 `App 2` 注册的 `request handler` 处理该请求。

在上述 **socket - http - app 三层模型** 中，对 socket 和 http，socket 是 Subject，http 是 Observer；对 http 和 app，http 是 Subject，app 是 Observer。

## 代码实现

因为在观察者模式的实现上，socket 模块和 http 模块类似，所以，下面只给出 socket 模块的实现：

```go
// demo/network/socket.go
package network

// 关键点1: 定义Observer接口
// SocketListener Socket报文监听者
type SocketListener interface {
  // 关键2: 为Observer定义更新处理方法，入参为相关的上下文对象
	Handle(packet *Packet) error
}

// Subject接口
// Socket 网络通信Socket接口
type Socket interface {
	// Listen 在endpoint指向地址上起监听
	Listen(endpoint Endpoint) error
	// Close 关闭监听
	Close(endpoint Endpoint)
	// Send 发送网络报文
	Send(packet *Packet) error
	// Receive 接收网络报文
	Receive(packet *Packet)
	// AddListener 增加网络报文监听者
	AddListener(listener SocketListener)
}

// 关键点3: 定义Subject对象
// socketImpl Socket的默认实现
type socketImpl struct {
  // 关键点4: 在Subject中持有Observer的集合
	listeners []SocketListener
}

// 关键点5: 为Subject定义注册Observer的方法
func (s *socketImpl) AddListener(listener SocketListener) {
	s.listeners = append(s.listeners, listener)
}

// 关键点6: 当Subject状态变更时，遍历Observers集合，调用它们的更新处理方法
func (s *socketImpl) Receive(packet *Packet) {
	for _, listener := range s.listeners {
		listener.Handle(packet)
	}
}

...
```

总结实现观察者模式的几个关键点：

1. 定义 Observer 接口，上述例子中为 `SocketListener` 接口。
2. 为 Observer 接口定义状态更新的处理方法，其中方法入参为相关的上下文对象。上述例子为 `Handle` 方法，上下文对象为 `Packet`。
3. 定义 Subject 对象，上述例子为 `socketImpl` 对象。当然，也可以先将 Subject 抽象为接口，比如上述例子中的 `Socket` 接口，但大多数情况下都不是必须的。
4. 在 Subject 对象中，持有 Observer 接口的集合，上述例子为 `listeners` 属性。**让 Subject 依赖 Observer 接口，能够使 Subject 与具体的 Observer 实现解耦，提升代码的可扩展性**。
5. 为 Subject 对象定义注册 Observer 的方法，上述例子为 `AddListener` 方法。
6. 当 Subject 状态变更时，遍历 Observer 集合，并调用它们的状态更变处理方法，上述例子为 `Receive` 方法。

## 扩展

### 发布-订阅模式

与观察者模式相近的，是**发布-订阅模式**（**Pub-Sub Pattern**），很多人会把两者等同，但它们之间还是有些差异。

从前文的观察者模式实现中，我们发现 Subject 持有 Observer 的引用，当状态变更时，Subject 直接调用 Observer 的更新处理方法完成通知。也就是，Subject 知道有哪些 Observer，也知道 Observer 的数量：

![](https://tva1.sinaimg.cn/large/e6c9d24egy1h4gz7fhuz1j212u0oi42e.jpg)

在发布-订阅模式中，我们将发布方称为 **Publisher**，订阅方称为 **Subscriber**，不同于观察者模式，Publisher 并不直接持有 Subscriber 引用，它们之间通常通过 **Broker** 来完成解耦。也即，Publisher 不知道有哪些 Subscriber，也不知道 Subscriber 的数量：

![](https://tva1.sinaimg.cn/large/e6c9d24egy1h4gz9mepafj21am0p2gqv.jpg)

**发布-订阅模式被广泛应用在消息中间件的实现上**，比如 Apache Kafka 基于 Topic 实现了发布-订阅模式，发布方称为 Producer，订阅方称为 Consumer。

下面，我们通过 [简单的分布式应用系统](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation/blob/main/docs/go_ractice_design_pattern__solid_principle.md)（示例代码工程）中的 mq 模块，展示一个简单的发布-订阅模式实现，在该实现中，我们将 Publisher 的 produce 方法和 Subscriber 的 consume 方法都合并到 Broker 中：

```go
// demo/mq/memory_mq.go

// 关键点1: 定义通信双方交互的消息，携带topic信息
// Message 消息队列中消息定义
type Message struct {
	topic   Topic
	payload string
}

// 关键点2: 定义Broker对象
// memoryMq 内存消息队列，通过channel实现
type memoryMq struct {
  // 关键点3: Broker中维持一个队列的map，其中key为topic，value为queue，go语言通常用chan实现。
	queues sync.Map // key为Topic，value为chan *Message，每个topic单独一个队列
}

// 关键点4: 为Broker定义Produce方法，根据消息中的topic选择对应的queue发布消息
func (m *memoryMq) Produce(message *Message) error {
	record, ok := m.queues.Load(message.Topic())
	if !ok {
		q := make(chan *Message, 10000)
		m.queues.Store(message.Topic(), q)
		record = q
	}
	queue, ok := record.(chan *Message)
	if !ok {
		return errors.New("model's type is not chan *Message")
	}
	queue <- message
	return nil
}

// 关键点5: 为Broker定义Consume方法，根据topic选择对应的queue消费消息
func (m *memoryMq) Consume(topic Topic) (*Message, error) {
	record, ok := m.queues.Load(topic)
	if !ok {
		q := make(chan *Message, 10000)
		m.queues.Store(topic, q)
		record = q
	}
	queue, ok := record.(chan *Message)
	if !ok {
		return nil, errors.New("model's type is not chan *Message")
	}
	return <-queue, nil
}
```

客户端使用时，直接调用 `memoryMq` 的 `Produce` 方法和 `Consume` 方法完成消息的生产和消费：

```go
// 发布方
func publisher() {
	msg := NewMessage("test", "hello world")
	err := MemoryMqInstance().Produce(msg)
	assert.Nil(t, err)
}

// 订阅方
func subscriber() {
	result, err := MemoryMqInstance().Consume("test")
	assert.Nil(err)
	assert.Equal(t, "hello world", result.payload)
}
```

总结实现发布-订阅模式的几个关键点：

1. 定义通信双方交互的消息，携带 topic 信息，上述例子为 `Message` 对象。
2. 定义 Broker 对象，Broker 是缓存消息的地方，上述例子为 `memoryMq` 对象。
3. 在 Broker 中维持一个队列的 map，其中 key 为 topic，value 为 queue，**go 语言通常用 chan 来实现 queue**，上述例子为 `queues` 属性。
4. 为 Broker 定义 produce 方法，根据消息中的 topic 选择对应的 queue 发布消息，上述例子为 `Produce` 方法。
5. 为 Broker 定义 consume 方法，根据 topic 选择对应的 queue 消费消息，上述例子为 `Consume` 方法。

### Push 模式 VS Pull 模式

实现观察者模式和发布-订阅模式时，都会涉及到 **Push 模式**或 **Pull 模式**的选取。所谓 Push 模式，指的是 Subject/Publisher 直接将消息推送给 Observer/Subscriber；所谓 Pull 模式，指的是 Observer/Subscriber 主动向 Subject/Publisher 拉取消息：

![](https://tva1.sinaimg.cn/large/e6c9d24egy1h4h0mnk1zlj21aa0mk78n.jpg)

**Push 模式和 Pull 模式的选择，取决于通信双方处理消息的速率大小**。

如果 Subject/Publisher 方生产消息的速率要比 Observer/Subscriber 方处理消息的速率小，可以选择 Push 模式，以求得更高效、及时的消息传递；相反，如果 Subject/Publisher 方产生消息的速率要大，就要选择 Pull 模式，由 Observer/Subscriber 方决定消息的消费速率，否则可能导致 Observer/Subscriber 崩溃。

Pull 模式有个缺点，如果当前无消息可处理，将导致 Observer/Subscriber 空轮询，可以采用类似 Kafka 的解决方案：**让 Observer/Subscriber 阻塞一定时长，让出 CPU，避免长期无效的 CPU 空转**。

## 典型应用场景

- 需要监听某个状态的变更，且在状态变更时，通知到监听者。
- **web 框架**。很多 web 框架都用了观察者模式，用户注册请求 handler 到框架，框架收到相应请求后，调用 handler 完成处理逻辑。
- **消息中间件**。如 Kafka、RocketMQ 等。

## 优缺点

### 优点

- 消息通信双方解耦。观察者模式通过依赖接口达到松耦合；发布-订阅模式则通过 Broker 达到解耦目的。

- 支持广播通信。

- 可基于 topic 来达到**指定消费某一类型消息**的目的。

### 缺点

- 通知 Observer/Subscriber 的顺序是不确定的，应用程序不应该依赖通知顺序来保证业务逻辑的正确性。
- 广播通信场景，需要 Observer/Subscriber 自己去判断是否需要处理该消息，否则容易导致 **unexpected update**。

## 与其他模式的关联

观察者模式和发布-订阅模式中的 Subject 和 Broker，通常都会使用 [单例模式](https://mp.weixin.qq.com/s/fzdxrhziPkSqM_RBTeSPKA) 来确保它们全局唯一。

### 文章配图

可以在 [用Keynote画出手绘风格的配图](https://mp.weixin.qq.com/s/-sYW-oa6KzTR9LNdMWCSnQ) 中找到文章的绘图方法。

> #### 参考
>
> [1] [【Go实现】实践GoF的23种设计模式：SOLID原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA), 元闰子
>
> [2] [【Go实现】实践GoF的23种设计模式：单例模式](https://mp.weixin.qq.com/s/fzdxrhziPkSqM_RBTeSPKA), 元闰子
>
> [3] [Design Patterns, Chapter 5. Behavioral Patterns](https://learning.oreilly.com/library/view/design-patterns-elements/0201633612/), GoF
>
> [4] [观察者模式](https://refactoringguru.cn/design-patterns/observer), refactoringguru.cn
>
> [5] [观察者模式 vs 发布订阅模式](https://zhuanlan.zhihu.com/p/51357583), 柳树
>
> 更多文章请关注微信公众号：**元闰子的邀请**

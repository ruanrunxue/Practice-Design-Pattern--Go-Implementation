> 上一篇：[【Go实现】实践GoF的23种设计模式：建造者模式](https://mp.weixin.qq.com/s/LHezb8zWEFR7mRseqO2B_Q)
>
> **简单的分布式应用系统**（示例代码工程）：[https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation)

## 简述

 **工厂方法模式**（Factory Method Pattern）跟上一篇讨论的建造者模式类似，都是**将对象创建的逻辑封装起来，为使用者提供一个简单易用的对象创建接口**。两者在应用场景上稍有区别，建造者模式常用于需要传递多个参数来进行实例化的场景；工厂方法模式常用于**不指定对象具体类型的情况下创建对象**的场景。

## UML 结构

![](https://tva1.sinaimg.cn/large/e6c9d24egy1h2f9p62qckj21dm0s8n3o.jpg)

## 代码实现

### 示例

在[简单的分布式应用系统](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation/blob/main/docs/go_ractice_design_pattern__solid_principle.md)（示例代码工程）中，我们设计了 `Sidecar`  边车模块， `Sidecar`  的作用是为了给原生的 `Socket` 增加额外的功能，比如流控、日志等。

![](https://tva1.sinaimg.cn/large/e6c9d24egy1gzn32jkkduj213g0o00xq.jpg)

  `Sidecar`  模块的设计运用了**装饰者模式**，修饰的是 `Socket` 。所以客户端其实是把 `Sidecar` 当成是 `Socket` 来使用了，比如：

```go
// demo/network/http/http_client.go
package http

// 创建一个新的HTTP客户端，以Socket接口作为入参
func NewClient(socket network.Socket, ip string) (*Client, error) {
  ... // 一些初始化逻辑
	return client, nil
}

// 使用NewClient时，我们可以传入Sidecar来给Http客户端附加额外的流控功能
client, err := http.NewClient(sidecar.NewFlowCtrlSidecar(network.DefaultSocket()), "192.168.0.1")
```

在服务消息中介中，每次收到上游服务的 HTTP 请求，都会调用 `http.NewClient` 来创建一个 HTTP 客户端，并通过它将请求转发给下游服务：

```go
type ServiceMediator struct {
  ...
	server *http.Server
}

// Forward 转发请求，请求URL为 /{serviceType}+ServiceUri 的形式，如/serviceA/api/v1/task
func (s *ServiceMediator) Forward(req *http.Request) *http.Response {
    ...
    // 发现下游服务的目的IP地址
    dest, err := s.discovery(svcType)
    // 创建HTTP客户端，硬编码sidecar.NewFlowCtrlSidecar(network.DefaultSocket())
    client, err := http.NewClient(sidecar.NewFlowCtrlSidecar(network.DefaultSocket()), s.localIp)
    // 通过HTTP客户端转发请求
    resp, err := client.Send(dest, forwardReq)
    ...
}
```

在上述实现中，我们在调用 `http.NewClient` 时把 `sidecar.NewFlowCtrlSidecar(network.DefaultSocket())` 硬编码进去了，那么如果以后要扩展 `Sidecar` ，就得修改这段代码逻辑，这违反了[开闭原则 OCP](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA)。

有经验的同学可能会想到，可以通过让 `ServiceMediator` 依赖 `Socket` 接口，在 `Forward` 方法调用 `http.NewClient` 时把 `Socket` 接口作为入参；然后在 `ServiceMediator`  初始化时，将具体类型的 `Sidecar` 注入到 `ServiceMediator` 中：

```go
type ServiceMediator struct {
  ...
	server *http.Server
  // 依赖Socket抽象接口
  socket network.Socket
}

// Forward 转发请求，请求URL为 /{serviceType}+ServiceUri 的形式，如/serviceA/api/v1/task
func (s *ServiceMediator) Forward(req *http.Request) *http.Response {
    ...
    // 发现下游服务的目的IP地址
    dest, err := s.discovery(svcType)
    // 创建HTTP客户端，将s.socket抽象接口作为入参
    client, err := http.NewClient(s.socket, s.localIp)
    // 通过HTTP客户端转发请求
    resp, err := client.Send(dest, forwardReq)
    ...
}

// 在ServiceMediator初始化时，将具体类型的Sidecar注入到ServiceMediator中
mediator := &ServiceMediator{
  socket: sidecar.NewFlowCtrlSidecar(network.DefaultSocket())
}
```

上述的修改，从原来依赖具体，改成了依赖抽象，符合了开闭原则。

但是， `Forward` 方法存在并发调用的场景，因此它希望每次被调用时都创建一个新的 `Socket/Sidecar` 来完成网络通信，否则就需要加锁来保证并发安全。而上述的修改会导致在 `ServiceMediator` 的生命周期内都使用同一个 `Socket/Sidecar`，显然不符合要求。

因此，我们需要一个方法，既能够满足开闭原则，而且在每次调用`Forward` 方法时也能够创建新的 `Socket/Sidecar` 实例。工厂方法模式恰好就能满足这两点要求，下面我们通过它来完成代码的优化。

### 实现

```go
// demo/sidecar/sidecar_factory.go

// 关键点1: 定义一个Sidecar工厂抽象接口
type Factory interface {
  // 关键点2: 工厂方法返回Socket抽象接口
	Create() network.Socket
}

// 关键点3: 按照需要实现具体的工厂


// demo/sidecar/raw_socket_sidecar_factory.go
// RawSocketFactory 只具备原生socket功能的sidecar，实现了Factory接口
type RawSocketFactory struct {
}
func (r RawSocketFactory) Create() network.Socket {
	return network.DefaultSocket()
}

// demo/sidecar/all_in_one_sidecar_factory.go
// AllInOneFactory 具备所有功能的sidecar工厂，实现了Factory接口
type AllInOneFactory struct {
	producer mq.Producible
}
func (a AllInOneFactory) Create() network.Socket {
	return NewAccessLogSidecar(NewFlowCtrlSidecar(network.DefaultSocket()), a.producer)
}
```

上述代码中，我们定义了一个工厂抽象接口 `Factory` ，并有了 2 个具体的实现 `RawSocketFactory` 和 `AllInOneFactory`。最后， `ServiceMediator` 依赖  `Factory` ，并在 `Forward` 方法中通过 `Factory` 来创建新的 `Socket/Sidecar` ：

```go
// demo/service/mediator/service_mediator.go

type ServiceMediator struct {
  ...
	server *http.Server
  // 关键点4: 客户端依赖Factory抽象接口
  sidecarFactory sidecar.Factory
}

// Forward 转发请求，请求URL为 /{serviceType}+ServiceUri 的形式，如/serviceA/api/v1/task
func (s *ServiceMediator) Forward(req *http.Request) *http.Response {
    ...
    // 发现下游服务的目的IP地址
    dest, err := s.discovery(svcType)
    // 创建HTTP客户端，调用sidecarFactory.Create()生成Socket作为入参
    client, err := http.NewClient(s.sidecarFactory.Create(), s.localIp)
    // 通过HTTP客户端转发请求
    resp, err := client.Send(dest, forwardReq)
    ...
}

// 关键点5: 在ServiceMediator初始化时，将具体类型的sidecar.Factory注入到ServiceMediator中
mediator := &ServiceMediator{
  sidecarFactory: &AllInOneFactory{}
  // sidecarFactory: &RawSocketFactory{}
}
```

下面总结实现工厂方法模式的几个关键点：

1. 定义一个工厂方法抽象接口，比如前文中的 `sidecar.Factory`。
2. 工厂方法中，返回需要创建的对象/接口，比如 `network.Socket`。其中，工厂方法通常命名为 `Create`。
3. 按照具体需要，定义工厂方法抽象接口的具体实现对象，比如 `RawSocketFactory` 和 `AllInOneFactory`。
4. 客户端使用时，依赖工厂方法抽象接口。
5. 在客户端初始化阶段，完成具体工厂对象的依赖注入。

## 扩展

### Go 风格的实现

前文的工厂方法模式实现，是非常典型的**面向对象风格**，下面我们给出一个更具 Go 风格的实现。

```go
// demo/sidecar/sidecar_factory_func.go

// 关键点1: 定义Sidecar工厂方法类型
type FactoryFunc func() network.Socket

// 关键点2: 按需定义具体的工厂方法实现，注意这里定义的是工厂方法的工厂方法，返回的是FactoryFunc工厂方法类型
func RawSocketFactoryFunc() FactoryFunc {
	return func() network.Socket {
		return network.DefaultSocket()
	}
}

func AllInOneFactoryFunc(producer mq.Producible) FactoryFunc {
	return func() network.Socket {
		return NewAccessLogSidecar(NewFlowCtrlSidecar(network.DefaultSocket()), producer)
	}
}

type ServiceMediator struct {
  ...
	server *http.Server
  // 关键点3: 客户端依赖FactoryFunc工厂方法类型
  sidecarFactoryFunc FactoryFunc
}

func (s *ServiceMediator) Forward(req *http.Request) *http.Response {
    ...
    dest, err := s.discovery(svcType)
    // 关键点4: 创建HTTP客户端，调用sidecarFactoryFunc()生成Socket作为入参
    client, err := http.NewClient(s.sidecarFactoryFunc(), s.localIp)
    resp, err := client.Send(dest, forwardReq)
    ...
}

// 关键点5: 在ServiceMediator初始化时，将具体类型的FactoryFunc注入到ServiceMediator中
mediator := &ServiceMediator{
  sidecarFactoryFunc: RawSocketFactoryFunc()
  // sidecarFactory: AllInOneFactoryFunc(producer)
}
```

上述的实现，利用了 Go 语言中**函数作为一等公民**的特点，少定义了几个 `interface` 和 `struct`，代码更加的简洁。

几个实现的关键点与面向对象风格的实现类似。值得注意的是 `关键点2` ，我们相当于定义了一个**工厂方法的工厂方法**，这么做是为了利用函数闭包的特点来**传递参数**。如果直接定义工厂方法，那么 `AllInOneFactoryFunc` 的实现是下面这样的，无法实现多态：

```go
// 并非FactoryFunc类型，无法实现多态
func AllInOneFactoryFunc(producer mq.Producible) network.Socket {
    return NewAccessLogSidecar(NewFlowCtrlSidecar(network.DefaultSocket()), producer)
}
```

### 简单工厂

工厂方法模式的另一个变种是**简单工厂**，它并不通过多态，而是通过简单的 `switch-case/if-else` 条件判断来决定创建哪种产品：

```go
// demo/sidecar/sidecar_simple_factory.go

// 关键点1: 定义sidecar类型
type Type uint8

// 关键点2: 按照需要定义sidecar具体类型
const (
	Raw Type = iota
	AllInOne
)

// 关键点3: 定义简单工厂对象
type SimpleFactory struct {
	producer mq.Producible
}

// 关键点4: 定义工厂方法，入参为sidecar类型，根据switch-case或者if-else来创建产品
func (s SimpleFactory) Create(sidecarType Type) network.Socket {
	switch sidecarType {
	case Raw:
		return network.DefaultSocket()
	case AllInOne:
		return NewAccessLogSidecar(NewFlowCtrlSidecar(network.DefaultSocket()), s.producer)
	default:
		return nil
	}
}

// 关键点5: 创建产品时传入具体的sidecar类型，比如sidecar.AllInOne
simpleFactory := &sidecar.SimpleFactory{producer: producer}
sidecar := simpleFactory.Create(sidecar.AllInOne)
```

### 静态工厂方法

静态工厂方法是 Java/C++ 的说法，主要用于替代构造函数来完成对象的实例化，能够让代码的可读性更好，而且起到了与客户端解耦的作用。比如 Java 的静态工厂方法实现如下：

```java
public class Packet {
    private final Endpoint src;
    private final Endpoint dest;
    private final Object payload;

    private Packet(Endpoint src, Endpoint dest, Object payload) {
        this.src = src;
        this.dest = dest;
        this.payload = payload;
    }

    // 静态工厂方法
    public static Packet of(Endpoint src, Endpoint dest, Object payload) {
        return new Packet(src, dest, payload);
    }
		...
}

// 用法
packet = Packet.of(src, dest, payload)
```

Go 中并没有**静态**一说，直接通过普通函数来完成对象的构造即可，比如：

```go
// demo/network/packet.go
type Packet struct {
	src     Endpoint
	dest    Endpoint
	payload interface{}
}

// 工厂方法
func NewPacket(src, dest Endpoint, payload interface{}) *Packet {
	return &Packet{
		src:     src,
		dest:    dest,
		payload: payload,
	}
}

// 用法
packet := NewPacket(src, dest, payload)
```

## 典型应用场景

1. **对象实例化逻辑较为复杂**时，可选择使用工厂方法模式/简单工厂/静态工厂方法来进行封装，为客户端提供一个易用的接口。
2. 如果**实例化的对象/接口涉及多种实现**，可以使用工厂方法模式实现多态。
3. **普通对象的创建，推荐使用静态工厂方法**，比直接的实例化（比如 `&Packet{src: src, dest: dest, payload: payload}`）具备更好的可读性和低耦合。

## 优缺点

### 优点

1. 代码的可读性更好。
2. 与客户端程序解耦，当实例化逻辑变更时，只需改动工厂方法即可，避免了霰弹式修改。

### 缺点

1. 引入工厂方法模式会新增一些对象/接口的定义，滥用会导致代码更加复杂。

## 与其他模式的关联

很多同学容易将工厂方法模式和**抽象工厂模式**混淆，抽象工厂模式主要运用在实例化“**产品族**”的场景，可以看成是工厂方法模式的一种演进。

> #### 参考
>
> [1] [【Go实现】实践GoF的23种设计模式：SOLID原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA), 元闰子
>
> [2] [Design Patterns, Chapter 3. Creational Patterns](https://learning.oreilly.com/library/view/design-patterns-elements/0201633612/), GoF
>
> [3] [Factory patterns in Go (Golang)](https://www.sohamkamani.com/golang/2018-06-20-golang-factory-patterns/), Soham Kamani
>
> [4] [工厂方法](https://zh.wikipedia.org/wiki/工厂方法), 维基百科
>
> 更多文章请关注微信公众号：**元闰子的邀请**
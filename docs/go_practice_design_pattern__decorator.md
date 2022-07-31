> 上一篇：[【Go实现】实践GoF的23种设计模式：原型模式](https://mp.weixin.qq.com/s/4GBZzH1Na0oqV_w6BFD7-w)
>
> **简单的分布式应用系统**（示例代码工程）：[https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation)

## 简介

我们经常会遇到“**给现有对象/模块新增功能**”的场景，比如 http router 的开发场景下，除了最基础的路由功能之外，我们常常还会加上如日志、鉴权、流控等 middleware。如果你查看框架的源码，就会发现 middleware 功能的实现用的就是**装饰者模式**（Decorator Pattern）。

GoF 给装饰者模式的定义如下：

> Decorators provide a flexible alternative to subclassing for extending functionality. Attach additional responsibilities to an object dynamically. 

简单来说，装饰者模式通过**组合**的方式，提供了**能够动态地给对象/模块扩展新功能**的能力。理论上，只要没有限制，它可以一直把功能叠加下去，具有很高的灵活性。

> 如果写过 Java，那么一定对 I/O Stream 体系不陌生，它是装饰者模式的经典用法，客户端程序可以动态地为原始的输入输出流添加功能，比如按字符串输入输出，加入缓冲等，使得整个 I/O Stream 体系具有很高的可扩展性和灵活性。

## UML 结构

![](https://tva1.sinaimg.cn/large/007S8ZIlgy1gigmr38eysj31c40kqx6p.jpg)

## 场景上下文

在[简单的分布式应用系统](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation/blob/main/docs/go_ractice_design_pattern__solid_principle.md)（示例代码工程）中，我们设计了 Sidecar 边车模块，它的用处主要是为了 1）方便扩展 `network.Socket` 的功能，如增加日志、流控等非业务功能；2）让这些附加功能对业务程序隐藏起来，也即业务程序只须关心看到 `network.Socket` 接口即可。

![](https://tva1.sinaimg.cn/large/e6c9d24egy1h3m37f6im9j21ge0qi0yd.jpg)

## 代码实现

Sidecar 的这个功能场景，很适合使用装饰者模式来实现，代码如下：

```go
// demo/network/socket.go
package network

// 关键点1: 定义被装饰的抽象接口
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

// 关键点2: 提供一个默认的基础实现
type socketImpl struct {
    listener SocketListener
}

func DefaultSocket() *socketImpl {
    return &socketImpl{}
}

func (s *socketImpl) Listen(endpoint Endpoint) error {
    return Instance().Listen(endpoint, s)
}
... // socketImpl的其他Socket实现方法


// demo/sidecar/flowctrl_sidecar.go
package sidecar

// 关键点3: 定义装饰器，实现被装饰的接口
// FlowCtrlSidecar HTTP接收端流控功能装饰器，自动拦截Socket接收报文，实现流控功能
type FlowCtrlSidecar struct {
  // 关键点4: 装饰器持有被装饰的抽象接口作为成员属性
    socket network.Socket
    ctx    *flowctrl.Context
}

// 关键点5: 对于需要扩展功能的方法，新增扩展功能
func (f *FlowCtrlSidecar) Receive(packet *network.Packet) {
    httpReq, ok := packet.Payload().(*http.Request)
    // 如果不是HTTP请求，则不做流控处理
    if !ok {
        f.socket.Receive(packet)
        return
    }
    // 流控后返回429 Too Many Request响应
    if !f.ctx.TryAccept() {
        httpResp := http.ResponseOfId(httpReq.ReqId()).
            AddStatusCode(http.StatusTooManyRequest).
            AddProblemDetails("enter flow ctrl state")
        f.socket.Send(network.NewPacket(packet.Dest(), packet.Src(), httpResp))
        return
    }
    f.socket.Receive(packet)
}

// 关键点6: 不需要扩展功能的方法，直接调用被装饰接口的原生方法即可
func (f *FlowCtrlSidecar) Close(endpoint network.Endpoint) {
    f.socket.Close(endpoint)
}
... // FlowCtrlSidecar的其他方法

// 关键点7: 定义装饰器的工厂方法，入参为被装饰接口
func NewFlowCtrlSidecar(socket network.Socket) *FlowCtrlSidecar {
    return &FlowCtrlSidecar{
        socket: socket,
        ctx:    flowctrl.NewContext(),
    }
}

// demo/sidecar/all_in_one_sidecar_factory.go
// 关键点8: 使用时，通过装饰器的工厂方法，把所有装饰器和被装饰者串联起来
func (a AllInOneFactory) Create() network.Socket {
    return NewAccessLogSidecar(NewFlowCtrlSidecar(network.DefaultSocket()), a.producer)
}
```

总结实现装饰者模式的几个关键点：

1. **定义需要被装饰的抽象接口**，后续的装饰器都是基于该接口进行扩展。
2. 为抽象接口提供一个基础实现。
3. 定义装饰器，并实现被装饰的抽象接口。
4. **装饰器持有被装饰的抽象接口作为成员属性**。“装饰”的意思是在原有功能的基础上扩展新功能，因此必须持有原有功能的抽象接口。
5. 在装饰器中，对于需要扩展功能的方法，新增扩展功能。
6. 不需要扩展功能的方法，**直接调用被装饰接口的原生方法即可**。
7. 为装饰器定义一个工厂方法，入参为被装饰接口。
8. 使用时，通过装饰器的工厂方法，把所有装饰器和被装饰者串联起来。

## 扩展

### Go 风格的实现

在 Sidecar 的场景上下文中，被装饰的 `Socket` 是一个相对复杂的接口，装饰器通过实现 `Socket` 接口来进行功能扩展，是典型的面向对象风格。

如果被装饰者是一个简单的接口/方法/函数，我们可以用更具 Go 风格的实现方式，考虑前文提到的 http router 场景。如果你使用原生的 `net/http` 进行 http router 开发，通常会这么实现：

```go
func main() {
  // 注册/hello的router
    http.HandleFunc("/hello", hello)
  // 启动http服务器
    http.ListenAndServe("localhost:8080", nil)
}

// 具体的请求处理逻辑，类型是 http.HandlerFunc
func hello(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello, world"))
}
```

其中，我们通过 `http.HandleFunc` 来注册具体的 router， `hello` 是具体的请求处理方法。现在，我们想为该 http 服务器增加日志、鉴权等通用功能，那么可以把 `func(w http.ResponseWriter, r *http.Request)` 作为被装饰的抽象接口，通过新增日志、鉴权等装饰器完成功能扩展。

```go
// demo/network/http/http_handle_func_decorator.go

// 关键点1: 确定被装饰接口，这里为原生的http.HandlerFunc
type HandlerFunc func(ResponseWriter, *Request)

// 关键点2: 定义装饰器类型，是一个函数类型，入参和返回值都是 http.HandlerFunc 函数
type HttpHandlerFuncDecorator func(http.HandlerFunc) http.HandlerFunc

// 关键点3: 定义装饰函数，入参为被装饰的接口和装饰器可变列表
func Decorate(h http.HandlerFunc, decorators ...HttpHandlerFuncDecorator) http.HandlerFunc {
    // 关键点4: 通过for循环遍历装饰器，完成对被装饰接口的装饰
    for _, decorator := range decorators {
        h = decorator(h)
    }
    return h
}

// 关键点5: 实现具体的装饰器
func WithBasicAuth(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("Auth")
        if err != nil || cookie.Value != "Pass" {
            w.WriteHeader(http.StatusForbidden)
            return
        }
        // 关键点6: 完成功能扩展之后，调用被装饰的方法，才能将所有装饰器和被装饰者串起来
        h(w, r)
    }
}

func WithLogger(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        log.Println(r.Form)
        log.Printf("path %s", r.URL.Path)
        h(w, r)
    }
}

func hello(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello, world"))
}

func main() {
    // 关键点7: 通过Decorate函数完成对hello的装饰
    http.HandleFunc("/hello", Decorate(hello, WithLogger, WithBasicAuth))
    // 启动http服务器
    http.ListenAndServe("localhost:8080", nil)
}
```

上述的装饰者模式的实现，用到了类似于 [Functional Options](https://mp.weixin.qq.com/s/LHezb8zWEFR7mRseqO2B_Q) 的技巧，也是巧妙利用了 Go 的函数式编程的特点，总结下来有如下几个关键点：

1. 确定被装饰的接口，上述例子为 `http.HandlerFunc`。
2. **定义装饰器类型，是一个函数类型，入参和返回值都是被装饰接口**，上述例子为 `func(http.HandlerFunc) http.HandlerFunc`。
3. 定义装饰函数，**入参为被装饰的接口和装饰器可变列表**，上述例子为 `Decorate` 方法。
4. 在装饰方法中，**通过for循环遍历装饰器，完成对被装饰接口的装饰**。这里是用来类似 [Functional Options](https://mp.weixin.qq.com/s/LHezb8zWEFR7mRseqO2B_Q) 的技巧，**一定要注意装饰器的顺序**！
5. 实现具体的装饰器，上述例子为 `WithBasicAuth` 和 `WithLogger` 函数。
6. 在装饰器中，完成功能扩展之后，记得调用被装饰者的接口，这样才能将所有装饰器和被装饰者串起来。
7. 在使用时，通过装饰函数完成对被装饰者的装饰，上述例子为 `Decorate(hello, WithLogger, WithBasicAuth)`。

### Go 标准库中的装饰者模式

在 Go 标准库中，也有一个运用了装饰者模式的模块，就是 `context`，其中关键的接口如下：

```go
package context

// 被装饰接口
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}

// cancel装饰器
type cancelCtx struct {
    Context // 被装饰接口
    mu       sync.Mutex
    done     atomic.Value
    children map[canceler]struct{}=
    err      error
}
// cancel装饰器的工厂方法
func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
    // ...  
    c := newCancelCtx(parent)
    propagateCancel(parent, &c)
    return &c, func() { c.cancel(true, Canceled) }
}

// timer装饰器
type timerCtx struct {
    cancelCtx // 被装饰接口
    timer *time.Timer

    deadline time.Time
}
// timer装饰器的工厂方法
func WithDeadline(parent Context, d time.Time) (Context, CancelFunc) {
  // ...
    c := &timerCtx{
        cancelCtx: newCancelCtx(parent),
        deadline:  d,
    }
    // ...
  return c, func() { c.cancel(true, Canceled) }
}
// timer装饰器的工厂方法
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
    return WithDeadline(parent, time.Now().Add(timeout))
}

// value装饰器
type valueCtx struct {
    Context // 被装饰接口
    key, val any
}
// value装饰器的工厂方法
func WithValue(parent Context, key, val any) Context {
    if parent == nil {
        panic("cannot create context from nil parent")
    }
  // ...
    return &valueCtx{parent, key, val}
}
```

![](https://tva1.sinaimg.cn/large/e6c9d24egy1h3ocsp168sj21dw0rs7bg.jpg)

使用时，可以这样：

```go
// 使用时，可以这样
func main() {
    ctx := context.Background()
    ctx = context.WithValue(ctx, "key1", "value1")
    ctx, _ = context.WithTimeout(ctx, time.Duration(1))
    ctx = context.WithValue(ctx, "key2", "value2")
}
```

不管是 UML 结构，还是使用方法，`context` 模块都与传统的装饰者模式有一定出入，但也不妨碍 `context` 是装饰者模式的典型运用。还是那句话，**学习设计模式，不能只记住它的结构，而是学习其中的动机和原理**。

## 典型使用场景

- **I/O 流**，比如为原始的 I/O 流增加缓冲、压缩等功能。
- **Http Router**，比如为基础的 Http Router 能力增加日志、鉴权、Cookie等功能。
- ......

## 优缺点

### 优点

1. 遵循[开闭原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA)，能够在不修改老代码的情况下扩展新功能。
2. 可以用多个装饰器把多个功能组合起来，理论上可以无限组合。

### 缺点

1. **一定要注意装饰器装饰的顺序**，否则容易出现不在预期内的行为。
2. 当装饰器越来越多之后，系统也会变得复杂。

## 与其他模式的关联

装饰者模式和**代理模式**具有很高的相似性，但是两种所强调的点不一样。**前者强调的是为本体对象添加新的功能；后者强调的是对本体对象的访问控制**。

装饰者模式和适配器模式的区别是，前者只会扩展功能而不会修改接口；后者则会修改接口。

### 文章配图

可以在 [用Keynote画出手绘风格的配图](https://mp.weixin.qq.com/s/-sYW-oa6KzTR9LNdMWCSnQ) 中找到文章的绘图方法。

> #### 参考
>
> [1] [【Go实现】实践GoF的23种设计模式：SOLID原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA), 元闰子
>
> [2] [【Go实现】实践GoF的23种设计模式：建造者模式](https://mp.weixin.qq.com/s/LHezb8zWEFR7mRseqO2B_Q), 元闰子
>
> [3] [Design Patterns, Chapter 4. Structural Patterns](https://learning.oreilly.com/library/view/design-patterns-elements/0201633612/), GoF
>
> [4] [装饰模式](https://refactoringguru.cn/design-patterns/decorator), refactoringguru.cn
>
> [5] [Golang Decorator Pattern](https://www.henrydu.com/2022/01/05/golang-decorator-pattern/), Henry Du
>
> 更多文章请关注微信公众号：**元闰子的邀请**

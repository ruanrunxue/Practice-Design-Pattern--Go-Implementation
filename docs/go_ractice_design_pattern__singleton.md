> 上一篇：[【Go实现】实践GoF的23种设计模式：SOLID原则](https://mp.weixin.qq.com/s?__biz=Mzg3MjAyNjUyMQ==&mid=2247484390&idx=1&sn=b9bc063243139d1df166a55feb6c6092&chksm=cef4db10f9835206c8ad9070b21b326766215df8d3accfff0d211c1e99105cf77e192548fdc6&token=415480090&lang=zh_CN#rd)
>
> **简单的分布式应用系统**（示例代码工程）：[https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation)

## 简述
GoF 对单例模式（Singleton）的定义如下：
> Ensure a class only has one instance, and provide a global point of access to it.

也即，**保证一个类只有一个实例，并且为它提供一个全局访问点**。

在程序设计中，有些对象通常只需要一个共享的实例，比如线程池、全局缓存、对象池等。实现共享实例最简单直接的方式就是**全局变量**。但是，使用全局变量会带来一些问题，比如：

1. 客户端程序可以创建同类实例，从而无法保证在整系统上只有一个共享实例。
1. 难以控制对象的访问，比如想增加一个“访问次数统计”的功能就很难，可扩展性较低。
1. 把实现细节暴露给客户端程序，加深了耦合，容易产生霰弹式修改。

对这种全局唯一的场景，更好的是使用单例模式去实现。**单例模式能够限制客户端程序创建同类实例，并且可以在全局访问点上扩展或修改功能，而不影响客户端程序**。

但是，并非所有的**全局唯一**都适用单例模式。比如下面这种场景：

> 考虑需要统计一个API调用的情况，有两个指标，成功调用次数和失败调用次数。这两个指标都是全局唯一的，所以有人可能会将其建模成两个单例SuccessApiMetric和FailApiMetric。按照这个思路，随着指标数量的增多，你会发现代码里类的定义会越来越多，也越来越臃肿。这也是单例模式最常见的误用场景，更好的方法是将两个指标设计成一个对象ApiMetric下的两个实例ApiMetic success和ApiMetic fail。

那么，如何判断一个对象是否应该被建模成单例？通常，被建模成单例的对象都有“**中心点**”的含义，比如线程池就是管理所有线程的中心。所以，**在判断一个对象是否适合单例模式时，先思考下，是一个中心点吗**？

## UML结构
![](https://tva1.sinaimg.cn/large/e6c9d24egy1h0z9d5373jj218q0iwn0g.jpg)
## 代码实现
根据单例模式的定义，实现的关键点有两个：

1. **限制调用者直接实例化该对象**；
1. **为该对象的单例提供一个全局唯一的访问方法**。

对于 C++ / Java 而言，只需把对象的构造函数设计成私有的，并提供一个 static 方法去访问该对象的唯一实例即可。但 Go 语言并没有构造函数的概念，也没有 static 方法，所以需要另寻出路。

我们可以利用 Go 语言 package 的访问规则来实现，将单例对象设计成首字母小写，这样就能限定它的访问范围只在当前`package`下，模拟了 C++ / Java 的私有构造函数；然后，在当前 package 下实现一个首字母大写的访问函数，也就相当于 static 方法的作用了。

### 示例
在[简单的分布式应用系统](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation/blob/main/docs/go_ractice_design_pattern__solid_principle.md)（示例代码工程）中，我们定义了一个网络模块 network，模拟实现了网络报文转发功能。network 的设计也很简单，通过一个哈希表维持了 `Endpoint` 到 `Socket` 的映射，报文转发时，通过 `Endpoint` 寻址到 `Socket`，再调用 `Socket` 的 `Receive` 方法完成转发。

![](https://tva1.sinaimg.cn/large/e6c9d24egy1h0yyva76hhj219i0lajw7.jpg)

因为整系统只需一个 network 对象，而且它在领域模型中具有**中心点**的语义，所以我们很自然地使用单例模式来实现它。单例模式大致可以分成两类，“饿汉模式”和“懒汉模式”。前者是在系统初始化期间就完成了单例对象的实例化；后者则是在调用时才进行延迟实例化，从而一定程度上节省了内存。

### “饿汉模式”实现
```go
// demo/network/network.go
package network

// 1、设计为小写字母开头，表示只在network包内可见，限制客户端程序的实例化
type network struct {
	sockets sync.Mapvar instancevar instance
}

// 2、定义一个包内可见的实例对象，也即单例
var instance = &network{sockets: sync.Map{}}

// 3、定义一个全局可见的唯一访问方法
func Instance() *network {
	return instance
}

func (n *network) Listen(endpoint Endpoint, socket Socket) error {
	if _, ok := n.sockets.Load(endpoint); ok {
		return ErrEndpointAlreadyListened
	}
	n.sockets.Store(endpoint, socket)
	return nil
}

func (n *network) Send(packet *Packet) error {
	record, rOk := n.sockets.Load(packet.Dest())
	socket, sOk := record.(Socket)
	if !rOk || !sOk {
		return ErrConnectionRefuse
	}
	go socket.Receive(packet)
	return nil
}
```
那么，客户端就可以通过 `network.Instance()` 引用该单例了：

```go
// demo/sidecar/flowctrl_sidecar.go
package sidecar

type FlowCtrlSidecar struct {...}

// 通过 network.Instance() 直接引用单例
func (f *FlowCtrlSidecar) Listen(endpoint network.Endpoint) error {
	return network.Instance().Listen(endpoint, f)
}
...
```
### “懒汉模式”实现
众所周知，“懒汉模式”会带来线程安全问题，可以通过**普通加锁**，或者更高效的**双重检验加锁**来优化。不管是哪种方法，都是为了**保证单例只会被初始化一次**。
```go
type network struct {...}

// 单例
var instance *network
// 定义互斥锁
var mutex = sync.Mutex{}

// 普通加锁，缺点是每次调用 Instance() 都需要加锁
func Instance() *network {
	mutex.Lock()
	if instance == nil {
		instance = &network{sockets: sync.Map{}}
	}
	mutex.Unlock()
	return instance
}

// 双重检验后加锁，实例化后无需加锁
func Instance() *network {
	if instance == nil {
        mutex.Lock()
        if instance == nil {
           instance = &network{sockets: sync.Map{}}
        }
        mutex.Unlock()
	}
	return instance
}
```
对于“懒汉模式”，Go 语言还有一个更优雅的实现方式，那就是利用 sync.Once。它有一个 Do 方法，方法声明为 `func (o *Once) Do(f func())`，其中入参是 `func()` 的方法类型，Go 会保证该方法仅会被调用一次。利用这个特性，我们就能够实现单例只被初始化一次了。
```go
type network struct {...}
// 单例
var instance *network
// 定义 once 对象
var once = sync.Once{}

// 通过once对象确保instance只被初始化一次
func Instance() *network {
	once.Do(func() {
        // 只会被调用一次
		instance = &network{sockets: sync.Map{}}
	})
	return instance
}
```
## 扩展
### 提供多个实例
虽然单例模式从定义上表示每个对象只能有一个实例，但是我们不应该被该定义限制住，还得从模式本身的动机来去理解它。单例模式的一大动机是**限制客户端程序对对象进行实例化**，至于实例有多少个其实并不重要，根据具体场景来进行建模、设计即可。

比如在前面的 network 模块中，现在新增一个这样的需求，将网络拆分为互联网和局域网。那么，我们可以这么设计：

```go
type network struct {...}

// 定义互联网单例
var inetInstance = &network{sockets: sync.Map{}}
// 定义局域网单例
var lanInstance = &network{sockets: sync.Map{}}


// 定义互联网全局可见的唯一访问方法
func Internet() *network {
	return inetInstance
}
// 定义局域网全局可见的唯一访问方法
func Lan() *network {
	return lanInstance
}
```
虽然上述例子中，`network` 结构有两个实例，但是本质上还是单例模式，因为它做到了限制客户端实例化，以及为每个单例提供了全局唯一的访问方法。
### 提供多种实现
单例模式也可以实现多态，如果你预测该单例未来可能会扩展，那么就可以将它设计成抽象的接口，**让客户端依赖抽象，这样，未来扩展时就无需改动客户端程序了**。

比如，我们可以 `network` 设计为一个抽象接口：

```go
// network 抽象接口
type network interface {
	Listen(endpoint Endpoint, socket Socket) error
	Send(packet *Packet) error
}

// network 的实现1
type networkImpl1 struct {
	sockets sync.Map
}
func (n *networkImpl1) Listen(endpoint Endpoint, socket Socket) error {...}
func (n *networkImpl1) Send(packet *Packet) error {...}

// networkImpl1 实现的单例
var instance = &networkImpl1{sockets: sync.Map{}}

// 定义全局可见的唯一访问方法，注意返回值时network抽象接口！
func Instance() network {
	return instance
}

// 客户端使用示例
func client() {
    packet := network.NewPacket(srcEndpoint, destEndpoint, payload)
    network.Instance().Send(packet)
}
```
如果未来需要新增一种 `networkImpl2` 实现，那么我们只需修改 `instance` 的初始化逻辑即可，客户端程序无需改动：
```go
// 新增network 的实现2
type networkImpl2 struct {...}
func (n *networkImpl2) Listen(endpoint Endpoint, socket Socket) error {...}
func (n *networkImpl2) Send(packet *Packet) error {...}

// 将单例 instance 修改为 networkImpl2 实现
var instance = &networkImpl2{...}

// 单例全局访问方法无需改动
func Instance() network {
	return instance
}

// 客户端使用也无需改动
func client() {
    packet := network.NewPacket(srcEndpoint, destEndpoint, payload)
    network.Instance().Send(packet)
}
```
有时候，我们还可能需要通过读取配置来决定使用哪种单例实现，那么，我们可以通过 `map` 来维护所有的实现，然后根据具体配置来选取对应的实现：
```go
// network 抽象接口
type network interface {
	Listen(endpoint Endpoint, socket Socket) error
	Send(packet *Packet) error
}

// network 具体实现
type networkImpl1 struct {...}
type networkImpl2 struct {...}
type networkImpl3 struct {...}
type networkImpl4 struct {...}

// 单例 map
var instances = make(map[string]network)

// 初始化所有的单例
func init() {
	instances["impl1"] = &networkImpl1{...}
	instances["impl2"] = &networkImpl2{...}
	instances["impl3"] = &networkImpl3{...}
	instances["impl4"] = &networkImpl4{...}
}

// 全局单例访问方法，通过读取配置决定使用哪种实现
func Instance() network {
    impl := readConf()
    instance, ok := instances[impl]
    if !ok {
        panic("instance not found")
    }
    return instance
}
```
## 典型应用场景

1. **日志**。每个服务通常都会需要一个全局的日志对象来记录本服务产生的日志。
1. **全局配置**。对于一些全局的配置，可以通过定义一个单例来供客户端使用。
1. **唯一序列号生成**。唯一序列号生成必然要求整系统只能有一个生成实例，非常合适使用单例模式。
1. **线程池、对象池、连接池等**。xxx池的本质就是**共享**，也是单例模式的常见场景。
1. **全局缓存**
1. ......
## 优缺点
### 优点
在合适的场景，使用单例模式有如下的**优点**：

1. 整系统只有一个或几个实例，有效节省了内存和对象创建的开销。
1. 通过全局访问点，可以方便地扩展功能，比如新增加访问次数的统计。
1. 对客户端隐藏实现细节，可避免霰弹式修改。
### 缺点
虽然单例模式相比全局变量有诸多的优点，但它本质上还是一个“全局变量”，还是避免不了全局变量的一些**缺点**：

1. **函数调用的隐式耦合**。通常我们都期望从函数的声明中就能知道该函数做了什么、依赖了什么、返回了什么。使用使用单例模式就意味着，无需通过函数传参，就能够在函数中使用该实例。也即将依赖/耦合隐式化了，不利于更好地理解代码。
1. **对测试不友好**。通常对一个方法/函数进行测试，我们并不需要知道它的具体实现。但如果方法/函数中有使用单例对象，我们就不得不考虑单例状态的变化了，也即需要考虑方法/函数的具体实现了。
1. **并发问题**。共享就意味着可能存在并发问题，我们不仅需要在初始化阶段考虑并发问题，在初始化后更是要时刻注意。因此，在高并发的场景，单例模式也可能存在锁冲突问题。

**_单例模式虽然简单易用，但也是最容易被滥用的设计模式。它并不是“银弹”，在实际使用时，还需根据具体的业务场景谨慎使用。_**
## 与其他模式的关联
**工厂方法模式**、**抽象工厂模式**很多时候都会以单例模式来实现，因为**工厂类**通常是无状态的，而且全局只需一个实例即可，能够有效避免对象的频繁创建和销毁。


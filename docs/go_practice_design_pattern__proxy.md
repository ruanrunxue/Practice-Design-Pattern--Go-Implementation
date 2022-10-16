> 上一篇：[【Go实现】实践GoF的23种设计模式：访问者模式](https://mp.weixin.qq.com/s/0qAIhbh4g_9Cde5tR2x24A)
>
> **简单的分布式应用系统**（示例代码工程）：[https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation)

## 简介

GoF 对代理模式（Proxy Pattern）的定义如下：

> Provide a surrogate or placeholder for another object to control access to it.

也即，**代理模式为一个对象提供一种代理以控制对该对象的访问**。

它是一个使用率非常高的设计模式，在现实生活中，也是很常见。比如，演唱会门票黄牛。假设你需要看一场演唱会，但官网上门票已经售罄，于是就当天到现场通过黄牛高价买了一张。在这个例子中，黄牛就相当于演唱会门票的代理，在正式渠道无法购买门票的情况下，你通过代理完成了该目标。

从演唱会门票的例子我们也能看出，使用代理模式的关键在于，**当 Client 不方便直接访问一个对象时，提供一个代理对象控制该对象的访问**。Client 实际上访问的是代理对象，代理对象会将 Client 的请求转给本体对象去处理。

## UML 结构

![](https://tva1.sinaimg.cn/large/008vxvgGgy1h76dqvxvedj313i0l2gov.jpg)

## 场景上下文

在 [简单的分布式应用系统](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation/blob/main/docs/go_ractice_design_pattern__solid_principle.md)（示例代码工程）中，db 模块用来存储服务注册和监控信息，它是一个 key-value 数据库。为了提升访问数据库的性能，我们决定为它新增一层缓存：

![](https://tva1.sinaimg.cn/large/008vxvgGgy1h76tsz2onhj317k0osgpx.jpg)

另外，我们希望客户端在使用数据库时，并不感知缓存的存在，这些，代理模式可以做到。

## 代码实现

```go
// demo/db/cache.go
package db

// 关键点1: 定义代理对象，实现被代理对象的接口
type CacheProxy struct {
  // 关键点2: 组合被代理对象，这里应该是抽象接口，提升可扩展性
    db    Db
    cache sync.Map // key为tableName，value为sync.Map[key: primaryId, value: interface{}]
    hit   int
    miss  int
}

// 关键点3: 在具体接口实现上，嵌入代理本身的逻辑
func (c *CacheProxy) Query(tableName string, primaryKey interface{}, result interface{}) error {
    cache, ok := c.cache.Load(tableName)
    if ok {
        if record, ok := cache.(*sync.Map).Load(primaryKey); ok {
            c.hit++
            result = record
            return nil
        }
    }
    c.miss++
    if err := c.db.Query(tableName, primaryKey, result); err != nil {
        return err
    }
    cache.(*sync.Map).Store(primaryKey, result)
    return nil
}

func (c *CacheProxy) Insert(tableName string, primaryKey interface{}, record interface{}) error {
    if err := c.db.Insert(tableName, primaryKey, record); err != nil {
        return err
    }
    cache, ok := c.cache.Load(tableName)
    if !ok {
        return nil
    }
    cache.(*sync.Map).Store(primaryKey, record)
    return nil
}

...

// 关键点4: 代理也可以有自己特有方法，提供一些辅助的功能
func (c *CacheProxy) Hit() int {
    return c.hit
}

func (c *CacheProxy) Miss() int {
    return c.miss
}

...
```

客户端这样使用：

```go
// 客户端只看到抽象的Db接口
func client(db Db) {
    table := NewTable("region").
      WithType(reflect.TypeOf(new(testRegion))).
      WithTableIteratorFactory(NewRandomTableIteratorFactory())
    db.CreateTable(table)
    table.Insert(1, &testRegion{Id: 1, Name: "region"})

    result := new(testRegion)
    db.Query("region", 1, result)
}

func main() {
    // 关键点5: 在初始化阶段，完成缓存的实例化，并依赖注入到客户端
    cache := NewCacheProxy(&memoryDb{tables: sync.Map{}})
    client(cache)
}
```

本例子中，Subject 是 `Db` 接口，Proxy 是 `CacheProxy` 对象，SubjectImpl 是 `memoryDb` 对象：

![](https://tva1.sinaimg.cn/large/008vxvgGgy1h76wlwt7v9j319g0pcjwd.jpg)

总结实现代理模式的几个关键点：

1. 定义代理对象，实现被代理对象的接口。本例子中，前者是 `CacheProxy` 对象，后者是 `Db` 接口。
2. 代理对象组合被代理对象，这里组合的应该是抽象接口，让代理的可扩展性更高些。本例子中，`CacheProxy` 对象组合了 `Db` 接口。
3. 代理对象在具体接口实现上，嵌入代理本身的逻辑。本例子中，`CacheProxy` 在 `Query`、`Insert` 等方法中，加入了缓存 `sync.Map` 的读写逻辑。
4. 代理对象也可以有自己特有方法，提供一些辅助的功能。本例子中，`CacheProxy` 新增了`Hit`、`Miss` 等方法用于统计缓存的命中率。
5. 最后，在初始化阶段，完成代理的实例化，并**依赖注入**到客户端。这要求，**客户端依赖抽象接口**，而不是具体实现，否则代理就不透明了。

## 扩展

### Go 标准库中的反向代理

代理模式最典型的应用场景是**远程代理**，其中，**反向代理**又是最常用的一种。

以 Web 应用为例，反向代理位于 Web 服务器前面，将客户端（例如 Web 浏览器）请求转发后端的 Web 服务器。**反向代理通常用于帮助提高安全性、性能和可靠性**，比如负载均衡、SSL 安全链接。

![](https://tva1.sinaimg.cn/large/008vxvgGgy1h7767wcipfj31bi0p6q8i.jpg)

Go 标准库的 net 包也提供了反向代理，`ReverseProxy`，位于 `net/http/httputil/reverseproxy.go` 下，实现 `http.Handler` 接口。`http.Handler` 提供了处理 Http 请求的能力，也即相当于 Http 服务器。那么，对应到 UML 结构图中，`http.Handler` 就是 Subject，`ReverseProxy` 就是 Proxy：

![](https://tva1.sinaimg.cn/large/008vxvgGgy1h776rhtwsij31ai0o8gqo.jpg)

下面列出 `ReverseProxy` 的一些核心代码：

```go
// net/http/httputil/reverseproxy.go
package httputil

type ReverseProxy struct {
    // 修改前端请求，然后通过Transport将修改后的请求转发给后端
    Director func(*http.Request)
    // 可理解为Subject，通过Transport来调用被代理对象的ServeHTTP方法处理请求
    Transport http.RoundTripper
    // 修改后端响应，并将修改后的响应返回给前端
    ModifyResponse func(*http.Response) error
    // 错误处理
    ErrorHandler func(http.ResponseWriter, *http.Request, error)
    ...
}

func (p *ReverseProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
    // 初始化transport
    transport := p.Transport
    if transport == nil {
        transport = http.DefaultTransport
    }
    ...
    // 修改前端请求
    p.Director(outreq)
    ...
    // 将请求转发给后端
    res, err := transport.RoundTrip(outreq)
    ...
    // 修改后端响应
    if !p.modifyResponse(rw, res, outreq) {
        return
    }
    ...
    // 给前端返回响应
    err = p.copyResponse(rw, res.Body, p.flushInterval(res))
    ...
}
```

`ReverseProxy` 就是典型的代理模式实现，其中，远程代理无法直接引用后端的对象引用，因此这里通过引入 `Transport` 来远程访问后端服务，可以将 `Transport`  理解为 Subject。

可以这么使用 `ReverseProxy`：

```go
func proxy(c *gin.Context) {
    remote, err := url.Parse("https://yrunz.com")
    if err != nil {
        panic(err)
    }

    proxy := httputil.NewSingleHostReverseProxy(remote)
    proxy.Director = func(req *http.Request) {
        req.Header = c.Request.Header
        req.Host = remote.Host
        req.URL.Scheme = remote.Scheme
        req.URL.Host = remote.Host
        req.URL.Path = c.Param("proxyPath")
    }

    proxy.ServeHTTP(c.Writer, c.Request)
}

func main() {
    r := gin.Default()
    r.Any("/*proxyPath", proxy)
    r.Run(":8080")
}
```

## 典型应用场景

- **远程代理**（remote proxy），远程代理适用于提供服务的对象处在远程的机器上，通过普通的函数调用无法使用服务，需要经过远程代理来完成。**因为并不能直接访问本体对象，所有远程代理对象通常不会直接持有本体对象的引用，而是持有远端机器的地址，通过网络协议去访问本体对象**。
- **虚拟代理**（virtual proxy），在程序设计中常常会有一些重量级的服务对象，如果一直持有该对象实例会非常消耗系统资源，这时可以通过虚拟代理来对该对象进行延迟初始化。
- **保护代理**（protection proxy），保护代理用于控制对本体对象的访问，常用于需要给 Client 的访问加上权限验证的场景。
- **缓存代理**（cache proxy），缓存代理主要在 Client 与本体对象之间加上一层缓存，用于加速本体对象的访问，常见于连接数据库的场景。
- **智能引用**（smart reference），智能引用为本体对象的访问提供了额外的动作，常见的实现为 C++ 中的智能指针，为对象的访问提供了计数功能，当访问对象的计数为 0 时销毁该对象。

## 优缺点

### 优点

- 可以在客户端不感知的情况下，控制访问对象，比如远程访问、增加缓存、安全等。
- 符合 [开闭原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA)，可以在不修改客户端和被代理对象的前提下，增加新的代理；也可以在不修改客户端和代理的前提下，更换被代理对象。

### 缺点

- 作为远程代理时，因为多了一次转发，会影响请求的时延。

## 与其他模式的关联

从结构上看，[装饰模式](https://mp.weixin.qq.com/s/NT6_KOY_hGkA-y2b4fw45A) 和 代理模式 具有很高的相似性，但是两种所强调的点不一样。**前者强调的是为本体对象添加新的功能，后者强调的是对本体对象的访问控制**。

### 文章配图

可以在 [用Keynote画出手绘风格的配图](https://mp.weixin.qq.com/s/-sYW-oa6KzTR9LNdMWCSnQ) 中找到文章的绘图方法。

> #### 参考
>
> [1] [【Go实现】实践GoF的23种设计模式：SOLID原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA), 元闰子
>
> [2] [【Go实现】实践GoF的23种设计模式：装饰模式](https://mp.weixin.qq.com/s/NT6_KOY_hGkA-y2b4fw45A), 元闰子
>
> [3] [Design Patterns, Chapter 4. Structural Patterns](https://learning.oreilly.com/library/view/design-patterns-elements/0201633612/ch04.html), GoF
>
> [4] [代理模式](https://refactoringguru.cn/design-patterns/proxy), refactoringguru.cn
>
> [5] [什么是反向代理？](https://www.cloudflare.com/zh-cn/learning/cdn/glossary/reverse-proxy/), cloudflare
>
> 更多文章请关注微信公众号：**元闰子的邀请**

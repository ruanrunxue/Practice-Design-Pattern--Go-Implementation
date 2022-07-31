> 上一篇：[【Go实现】实践GoF的23种设计模式：抽象工厂模式](https://mp.weixin.qq.com/s/RqqE6f3N_CzEWjdKluZhHg)
>
> **简单的分布式应用系统**（示例代码工程）：[https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation)

## 简介

**原型模式**（Prototype Pattern）主要解决**对象复制**的问题，它的核心就是 `Clone()` 方法，返回原型对象的复制品。

最简单直接的对象复制方式是这样的：重新实例化一个该对象的实例，然后遍历原始对象的所有成员变量， 并将成员变量值复制到新实例中。但这种方式的缺点也很明显：

1. 客户端程序必须清楚对象的实现细节。暴露细节往往不是件好事，它会导致代码耦合过深。
2. 对象可能存在一些私有属性，客户端程序无法访问它们，也就无法复制。
3. 很难保证所有的客户端程序都能完整不漏地把所有成员属性复制完。

更好的方法是使用原型模式，**将复制逻辑委托给对象本身**，这样，上述两个问题也都解决了。

### UML 结构

![](https://tva1.sinaimg.cn/large/007S8ZIlgy1ghky39ichjj319u0gqhdt.jpg)

## 场景上下文

在[简单的分布式应用系统](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation/blob/main/docs/go_ractice_design_pattern__solid_principle.md)（示例代码工程）中，我们设计了一个服务消息中介（Service Mediator）服务，可以把它看成是一个消息路由器，负责服务发现和消息转发：

![](https://tva1.sinaimg.cn/large/e6c9d24egy1gzn32jkkduj213g0o00xq.jpg)

消息转发也就意味着它必须**将上游服务的请求原封不动地转发给下游服务**，这是一个典型的对象复制场景。不过，在我们的实现里，服务消息中介会先修改上行请求的 URI，之后再转发给下游服务。因为上行请求 URI 中携带了下游服务的类型信息，用来做服务发现，在转发给下游服务时必须剔除。

比如，订单服务（order service）要发请求给库存服务（stock service），那么：

1. 订单服务先往服务消息中介发出 HTTP 请求，其中 URI 为 `/stock-service/api/v1/stock`。
2. 服务消息中介收到上行请求后，会从 URI 中提取出下游服务类型 `stock-service` ，通过服务注册中心发现库存服务的 Endpoint。
3. 随后，服务消息中介将修改后的请求转发给库存服务，其中 URI 为 `/api/v1/stock`。

![](https://tva1.sinaimg.cn/large/e6c9d24egy1h2u6m6f4xqj21f20t2ai0.jpg)

## 代码实现

如果按照简单直接的对象复制方式，实现是这样的：

```go
// 服务消息中介
type ServiceMediator struct {
    registryEndpoint network.Endpoint
    localIp          string
    server           *http.Server
    sidecarFactory   sidecar.Factory
}

// Forward 转发请求，请求URL为 /{serviceType}+ServiceUri 的形式，如/serviceA/api/v1/task
func (s *ServiceMediator) Forward(req *http.Request) *http.Response {
    // 提取上行请求URI中的服务类型
    svcType := s.svcTypeOf(req.Uri())
    // 剔除服务类型之后的请求URI
    svcUri := s.svcUriOf(req.Uri())
    // 根据服务类型做服务发现
    dest, err := s.discovery(svcType)
    if err != nil {
        ... // 异常处理
    }
    // 复制上行请求，将URI更改为剔除服务类型之后的URI
    forwardReq := http.EmptyRequest().
        AddUri(svcUri).
        AddMethod(req.Method()).
        AddHeaders(req.Headers()).
        AddQueryParams(req.QueryParams()).
        AddBody(req.Body())

    // 转发请求给下游服务  
    client, err := http.NewClient(s.sidecarFactory.Create(), s.localIp)
    if err != nil {
        ... // 异常处理
    }
    defer client.Close()
    resp, err := client.Send(dest, forwardReq)
    if err != nil {
        ... // 异常处理
    }
    
    // 复制下行响应，将ReqId更改为上行请求的ReqId，其他保持不变
    return http.NewResponse(req.ReqId()).
        AddHeaders(resp.Headers()).
        AddStatusCode(resp.StatusCode()).
        AddProblemDetails(resp.ProblemDetails()).
        AddBody(resp.Body())
}
...
```

上述实现中有 2 处进行了对象的复制：上行请求的复制和下行响应的复制。且不说直接进行对象复制具有前文提到的 3 种缺点，就代码可读性上来看也是稍显冗余。下面，我们使用原型模式进行优化。

首先，为 `http.Request` 和 `http.Response` 定义 `Clone` 方法：

```go
// demo/network/http/http_request.go
package http

type Request struct {
    reqId       ReqId
    method      Method
    uri         Uri
    queryParams map[string]string
    headers     map[string]string
    body        interface{}
}

// 关键点1: 定义原型复制方法Clone
func (r *Request) Clone() *Request {
  // reqId重新生成，其他都拷贝原来的值
    reqId := rand.Uint32() % 10000
    return &Request{
        reqId:       ReqId(reqId),
        method:      r.method,
        uri:         r.uri,
        queryParams: r.queryParams,
        headers:     r.headers,
        body:        r.body,
    }
}
...

// demo/network/http/http_response.go

type Response struct {
    reqId          ReqId
    statusCode     StatusCode
    headers        map[string]string
    body           interface{}
    problemDetails string
}

func (r *Response) Clone() *Response {
    return &Response{
        reqId:          r.reqId,
        statusCode:     r.statusCode,
        headers:        r.headers,
        body:           r.body,
        problemDetails: r.problemDetails,
    }
}
...
```

最后，在客户端程序处通过 `Clone` 方法来完成对象的复制：

```go
// demo/service/mediator/service_mediator.go

type ServiceMediator struct {...}

func (s *ServiceMediator) Forward(req *http.Request) *http.Response {
    ...
    dest, err := s.discovery(svcType)
    if err != nil {
        ...
    }
    // 关键点2: 通过Clone方法完成对象的复制，然后在此基础上进行进一步的修改
    forwardReq := req.Clone().AddUri(svcUri)
    ...
    resp, err := client.Send(dest, forwardReq)
    if err != nil {
        ...
    }
    return resp.Clone().AddReqId(req.ReqId())
}
```

原型模式的实现相对简单，可总结为 2 个关键点：

1. 为原型对象定义 `Clone` 方法，在此方法上完成成员属性的拷贝。
2. 在客户端程序中通过 `Clone` 来完成对象的复制。

需要注意的是，我们**不一定非得遵循标准的原型模式 UML 结构定义一个原型接口**，然后让原型对象实现它，比如：

```go
// Cloneable 原型复制接口
type Cloneable interface {
    Clone() Cloneable
}

type Response struct {...}
// 实现原型复制接口
func (r *Response) Clone() Cloneable {
    return &Response{
        reqId:          r.reqId,
        statusCode:     r.statusCode,
        headers:        r.headers,
        body:           r.body,
        problemDetails: r.problemDetails,
    }
}
```

在当前场景下，这样并不会给程序带来任何好处，反而新增一次类型强转，让程序变得更复杂了：

```go
func (s *ServiceMediator) Forward(req *http.Request) *http.Response {
    ...
    resp, err := client.Send(dest, forwardReq)
    if err != nil {
        ...
    }
    // 因为Clone方法返回的是Cloneable接口，因此需要转型为*http.Response
    return resp.Clone().(*http.Response).AddReqId(req.ReqId())
}
```

**所以，运用设计模式，最重要的是学得其中精髓，而不是仿照其形式，否则很容易适得其反**。

## 扩展

### 原型模式和与建造者模式的结合

原型模式和建造者模式相结合，也是常见的场景。还是以 `http.Request` 为例：

首先，我们先为它新增一个 `requestBuilder` 对象来完成对象的构造：

```go
// demo/network/http/http_request_builder.go
type requestBuilder struct {
    req *Request
}
// 普通Builder工厂方法，新创建一个Request对象
func NewRequestBuilder() *requestBuilder {
    return &requestBuilder{req: EmptyRequest()}
}

func (r *requestBuilder) AddMethod(method Method) *requestBuilder {
    r.req.method = method
    return r
}

func (r *requestBuilder) AddUri(uri Uri) *requestBuilder {
    r.req.uri = uri
    return r
}

... // 一系列 Addxxx 方法

func (r *requestBuilder) Builder() *Request {
    return r.req
}
```

下面，我们为 `requestBuilder` 新增一个 `NewRequestBuilderCopyFrom` 工厂方法来达到原型复制的效果：

```go
// demo/network/http/http_request_builder.go

// 实现原型模式的Builder工厂方法，复制已有的Request对象
func NewRequestBuilderCopyFrom(req *Request) *requestBuilder {
    reqId := rand.Uint32() % 10000
    replica := &Request{
        reqId:       ReqId(reqId),
        method:      req.method,
        uri:         req.uri,
        queryParams: req.queryParams,
        headers:     req.headers,
        body:        req.body,
    }
  // 将复制后的对象赋值给requestBuilder
    return &requestBuilder{req: replica}
}
```

用法如下：

```go
func (s *ServiceMediator) Forward(req *http.Request) *http.Response {
    ...
    dest, err := s.discovery(svcType)
    if err != nil {
        ...
    }
    // 原型模式和建造者模式相结合的实现
    forwardReq := http.NewRequestBuilderCopyFrom(req).Builder().AddUri(svcUri)
    ...
    resp, err := client.Send(dest, forwardReq)
    if err != nil {
        ...
    }
    // 普通原型模式的实现
    return resp.Clone().AddReqId(req.ReqId())
}
```

### 浅拷贝和深拷贝

如果原型对象的成员属性包含了指针类型，那么就会存在浅拷贝和深拷贝两种复制方式，比如对于原型对象 `ServiceProfile`，其中的 `Region` 属性为指针类型：

```go
// demo/service/registry/model/service_profile.go
package model

// ServiceProfile 服务档案，其中服务ID唯一标识一个服务实例，一种服务类型可以有多个服务实例
type ServiceProfile struct {
    Id       string           // 服务ID
    Type     ServiceType      // 服务类型
    Status   ServiceStatus    // 服务状态
    Endpoint network.Endpoint // 服务Endpoint
    Region   *Region          // 服务所属region
    Priority int              // 服务优先级，范围0～100，值越低，优先级越高
    Load     int              // 服务负载，负载越高表示服务处理的业务压力越大
}
```

浅拷贝的做法是直接复制指针：

```go
// 浅拷贝实现
func (s *ServiceProfile) Clone() Cloneable {
    return &ServiceProfile{
        Id:       s.Id,
        Type:     s.Type,
        Status:   s.Status,
        Endpoint: s.Endpoint,
        Region:   s.Region, // 指针复制，浅拷贝
        Priority: s.Priority,
        Load:     s.Load,
    }
}
```

深拷贝的做法则是创建新的 `Region` 对象：

```go
// 深拷贝实现
func (s *ServiceProfile) Clone() Cloneable {
    return &ServiceProfile{
        Id:       s.Id,
        Type:     s.Type,
        Status:   s.Status,
        Endpoint: s.Endpoint,
        Region: &Region{ // 新创建一个Region对象，深拷贝
            Id:      s.Region.Id,
            Name:    s.Region.Name,
            Country: s.Region.Country,
        },
        Priority: s.Priority,
        Load:     s.Load,
    }
}
```

具体使用哪种方式，因不同业务场景而异。浅拷贝直接复制指针，在性能上会好点；但某些场景下，引用同一个对象实例可能会导致业务异常，这时候就必须使用深拷贝了。

## 典型使用场景

1. 不管是复杂还是简单的对象，只要存在对象复制的场景，都适合使用原型模式。

## 优缺点

### 优点

1. 对客户端隐藏实现细节，有利于避免代码耦合。
2. 让客户端代码更简洁，有利于提升可读性。
3. 可方便地复制复杂对象，有利于杜绝客户端复制对象时的低级错误，比如漏复制属性。

### 缺点

1. 某些业务场景需要警惕浅拷贝问题。

## 与其他模式的关联

如前文提到的，原型模式和建造者模式相结合也是一种常见的应用场景。

> #### 参考
>
> [1] [【Go实现】实践GoF的23种设计模式：SOLID原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA), 元闰子
>
> [2] [【Go实现】实践GoF的23种设计模式：建造者模式](https://mp.weixin.qq.com/s/LHezb8zWEFR7mRseqO2B_Q), 元闰子
>
> [3] [Design Patterns, Chapter 3. Creational Patterns](https://learning.oreilly.com/library/view/design-patterns-elements/0201633612/), GoF
>
> 更多文章请关注微信公众号：**元闰子的邀请**


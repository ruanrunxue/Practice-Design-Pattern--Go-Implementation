> 上一篇：[【Go实现】实践GoF的23种设计模式：工厂方法模式](https://mp.weixin.qq.com/s/PwHc31ANLDVMNiagtqucZQ)
>
> **简单的分布式应用系统**（示例代码工程）：[https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation)

## 简述

上一篇我们介绍了[工厂方法模式](https://mp.weixin.qq.com/s/PwHc31ANLDVMNiagtqucZQ)，本文，我们继续介绍它的兄弟，**抽象工厂模式**（Abstract Factory Pattern）。

在工厂方法模式中，我们通过一个工厂方法接口来创建产品，而创建哪类产品，由具体的工厂对象来决定。抽象工厂模式和工厂方法模式的功能很类似，只是把“产品”，变成了“**产品族**”。**产品族就意味着这是一系列有关联的、一起使用的对象**。我们当然也可以为产品族中的每个产品定义一个工厂方法接口，但这显得有些冗余，因为一起使用通常也意味着同时创建，所以把它们放到同一个抽象工厂来创建会更合适。

## UML 结构

![](https://tva1.sinaimg.cn/large/e6c9d24egy1h2p1g5hgs1j21g40tqqan.jpg)

## 场景上下文

在[简单的分布式应用系统](https://github.com/ruanrunxue/Practice-Design-Pattern--Go-Implementation/blob/main/docs/go_ractice_design_pattern__solid_principle.md)（示例代码工程）中，我们有一个 Monitor 监控系统模块，该模块可以看成是一个简单的 ETL 系统，负责对监控数据的采集、处理、输出。整个模块被设计为插件化风格的架构，`Pipeline`是数据处理的流水线，其中包含了 `Input`、`Filter` 和 `Output` 三类插件，`Input` 负责从各类数据源中获取监控数据，`Filter` 负责数据处理，`Output` 负责将处理后的数据输出。更详细的设计思想我们在**桥接模式**一篇再做介绍，本文主要聚焦如何使用抽象工厂模式来解决各类插件的配置加载问题。

![](https://tva1.sinaimg.cn/large/e6c9d24egy1h2p1x43nj8j21ji0riwlf.jpg)

作为 ETL 系统，Monitor 模块应该具备灵活的扩展能力来应对不同的监控数据类型，因此，我们希望能够通过配置文件来定义 `Pipeline` 的行为。比如，下面就是一个 yaml 格式的配置内容：

```yaml
name: pipeline_0 # pipeline名称
type: simple # pipeline类型
input: # input插件定义
  name: input_0 # input插件名称
  type: memory_mq # input插件类型，这里使用的是MemoryMQ作为输入
  context: # input插件的配置上下文
    topic: access_log.topic # 这里配置的是订阅的MemoryMQ主题
filters: # filter插件链定义，多个filter插件组成一个filters插件链
  - name: filter_0 # filter插件名称
    type: extract_log # filter插件类型
  - name: filter_1
    type: add_timestamp
output: # output插件定义
  name: output_0 # output插件名称
  type: memory_db # output插件类型，这里使用的是MemoryDB作为输出
  context: # output插件上下文
    tableName: monitor_record_0 # 这里配置的是MemoryDB表名
```

另外，我们也希望 Monitor 模块支持多种类型的配置文件格式，比如，json 配置内容应该也支持：

```json
{
  "name": "pipeline_0",
  "type": "simple",
  "input": {
    "name": "input_0",
    "type": "memory_mq",
    "context": {
      "topic": "access_log.topic"
    }
  },
  "filters": [
    {
      "name": "filter_0",
      "type": "extract_log"
    },
    {
      "name": "filter_1",
      "type": "add_timestamp"
    }
  ],
  "output": {
    "name": "output_0",
    "type": "memory_db",
    "context": {
      "tableName": "monitor_record_0"
    }
  }
}
```

所以，整体的效果是这样的：

![](https://tva1.sinaimg.cn/large/e6c9d24egy1h2p37lnavjj21he0twwmw.jpg) 

可以看出，配置管理子模块中对象之间的关系，很符合抽象工厂模式的 UML 的结构，其中产品族就是 4 个插件配置对象，`conf.Input`、`conf.Filter`、`conf.Output`、`conf.Pipeline`，因此，我们下面使用抽象工厂模式来实现该子模块。

## 代码实现

首先，我们先把各个配置对象（产品）定义好：

```go
// demo/monitor/config/config.go
package config

// 配置基础结构
type item struct {
    Name       string         `json:"name" yaml:"name"`
    PluginType string         `json:"type" yaml:"type"`
    Ctx        plugin.Context `json:"context" yaml:"context"`
    loadConf   func(conf string, item interface{}) error // 封装不同配置文件的加载逻辑，实现多态的关键
}

// Input配置对象
type Input item

func (i *Input) Load(conf string) error {
    return i.loadConf(conf, i)
}

// Filter配置对象
type Filter item

func (f *Filter) Load(conf string) error {
    return f.loadConf(conf, f)
}

// Output配置对象
type Output item

func (o *Output) Load(conf string) error {
    return o.loadConf(conf, o)
}

// Pipeline配置对象
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

在 Java/C++ 等面向对象的编程语言中，我们定义一个产品的不同实现的时，通常采用继承的方式，也即先定义一个基类封装好公共逻辑，再定义不同的继承自该基类的不同子类来实现具体的逻辑。比如，对于 `Input` 配置对象，在 Java 中可能是这样定义的：

```java
// 基类
public abstract class InputConfig implements Config {
    protected String name;
    protected InputType type;
    protected Context ctx;

    // 子类实现具体加载逻辑
    @Override
    public abstract void load(String conf);
    ...
}
// Json子类
public class JsonInputConfig extends InputConfig {
    @Override
    public void load(String conf) {
        ... // Json配置文件加载逻辑
    }
}
// yaml子类
public class YamlInputConfig extends InputConfig {
    @Override
    public void load(String conf) {
        ... // Yaml配置文件加载逻辑
    }
}
```

但是在 Go 语言中并没有**继承**的概念，也无法定义抽象基类，因此，我们通过**定义一个函数对象 `loadConf` 来实现多态**，它的类型是 `func(conf string, item interface{}) error`，具体做的事情就是解析 `conf` 字符串（配置文件内容），然后完成 `item` 的赋值。

> Go 语言中通过函数对象来实现多态的技巧，我们在介绍**模板方法模式**时也会用到。

接下来，我们定义抽象工厂接口：

```go
// demo/monitor/config/config_factory.go

// 关键点1: 定义抽象工厂接口，里面定义了产品族中各个产品的工厂方法
type Factory interface {
    CreateInputConfig() Input
    CreateFilterConfig() Filter
    CreateOutputConfig() Output
    CreatePipelineConfig() Pipeline
}
```

然后是不同的实现：

```go
// demo/monitor/config/json_config_factory.go

// loadJson 加载json配置
func loadJson(conf string, item interface{}) error {
   return json.Unmarshal([]byte(conf), item)
}

// 关键点2: 实现抽象工厂接口
type JsonFactory struct {}

func NewJsonFactory() *JsonFactory {
    return &JsonFactory{}
}

// CreateInputConfig 例子 {"name":"input1", "type":"memory_mq", "context":{"topic":"monitor",...}}
func (j JsonFactory) CreateInputConfig() Input {
    return Input{loadConf: loadJson}
}

// CreateFilterConfig 例子 [{"name":"filter1", "type":"to_json"},{"name":"filter2", "type":"add_timestamp"},...]
func (j JsonFactory) CreateFilterConfig() Filter {
    return Filter{loadConf: loadJson}
}

// CreateOutputConfig 例子 {"name":"output1", "type":"memory_db", "context":{"tableName":"test",...}}
func (j JsonFactory) CreateOutputConfig() Output {
    return Output{loadConf: loadJson}
}

// CreatePipelineConfig 例子 {"name":"pipline1", "type":"simple", "input":{...}, "filter":{...}, "output":{...}}
func (j JsonFactory) CreatePipelineConfig() Pipeline {
    pipeline := Pipeline{}
    pipeline.loadConf = loadJson
    return pipeline
}


// demo/monitor/config/yaml_config_factory.go
// loadYaml 加载yaml配置
func loadYaml(conf string, item interface{}) error {
    return yaml.Unmarshal([]byte(conf), item)
}

// YamlFactory Yaml配置工厂
type YamlFactory struct {
}

func NewYamlFactory() *YamlFactory {
    return &YamlFactory{}
}

func (y YamlFactory) CreateInputConfig() Input {
    return Input{loadConf: loadYaml}
}

func (y YamlFactory) CreateFilterConfig() Filter {
    return Filter{loadConf: loadYaml}
}

func (y YamlFactory) CreateOutputConfig() Output {
    return Output{loadConf: loadYaml}
}

func (y YamlFactory) CreatePipelineConfig() Pipeline {
    pipeline := Pipeline{}
    pipeline.loadConf = loadYaml
    return pipeline
}
```

使用方法如下；

```go
// demo/monitor/monitor_system.go
type System struct {
    plugins       map[string]plugin.Plugin
    // 关键点3: 在使用时依赖抽象工厂接口
    configFactory config.Factory
}

func NewSystem(configFactory config.Factory) *System {
    return &System{
      plugins:       make(map[string]plugin.Plugin),
      configFactory: configFactory,
    }
}

func (s *System) LoadConf(conf string) error {
    pipelineConf := s.configFactory.CreatePipelineConfig()
    if err := pipelineConf.Load(conf); err != nil {
      return err
    }
    ...
}


// demo/example.go
func main() {
    // 关键点4: 在初始化是依赖注入具体的工厂实现
    monitorSys := monitor.NewSystem(config.NewYamlFactory())
    conf, _ := ioutil.ReadFile("monitor_pipeline.yaml")
    monitorSys.LoadConf(string(conf))
    ...
}
```

总结实现抽象工厂模式的几个关键点：

1. 定义抽象工厂接口，里面包含创建各个产品的工厂方法定义。
2. 定义抽象工厂接口的实现类。
3. 在客户端程序中依赖抽象工厂接口，通过接口来完成产品的创建。
4. 在客户端程序初始化时，将抽象工厂接口的具体实现依赖注入进去。

## 典型应用场景

1. 系统中有产品族，产品有不同的实现，且需要支持扩展。
2. 希望产品的创建逻辑和业务逻辑分离。

## 优缺点

### 优点

1. 产品创建逻辑和业务逻辑分离，符合单一职责原理。
2. 具有较高的可扩展性，新增一种产品族实现，只需新增一个抽象工厂实现即可。

### 缺点

1. 新增一些对象/接口的定义，滥用会导致代码更加复杂。

## 与其他模式的关联

很多同学容易将工厂方法模式和抽象工厂模式混淆，工厂方法模式主要应用在单个产品的实例化场景；抽象工厂模式则应用在“**产品族**”的实例化场景，可以看成是工厂方法模式的一种演进。

另外，抽象工厂接口的实现类，有时也会通过**单例模式**来实现。

> #### 参考
>
> [1] [【Go实现】实践GoF的23种设计模式：SOLID原则](https://mp.weixin.qq.com/s/s3aD4mK2Aw4v99tbCIe9HA), 元闰子
>
> [2] [【Go实现】实践GoF的23种设计模式：工厂方法模式](https://mp.weixin.qq.com/s/PwHc31ANLDVMNiagtqucZQ), 元闰子
>
> [3] [Design Patterns, Chapter 3. Creational Patterns](https://learning.oreilly.com/library/view/design-patterns-elements/0201633612/), GoF
>
> 更多文章请关注微信公众号：**元闰子的邀请**
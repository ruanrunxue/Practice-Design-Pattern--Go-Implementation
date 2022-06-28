package http

import (
    "context"
    "log"
    "net/http"
    "time"
)

// 关键点1: 确定被装饰者接口，这里为原生的http.HandlerFunc
// type HandlerFunc func(ResponseWriter, *Request)

// HttpHandlerFuncDecorator
// 关键点2: 定义装饰器类型，是一个函数类型，入参和返回值都是 http.HandlerFunc 函数
type HttpHandlerFuncDecorator func(http.HandlerFunc) http.HandlerFunc

// Decorate
// 关键点3: 定义装饰方法，入参为被装饰的接口和装饰器可变列表
func Decorate(h http.HandlerFunc, decorators ...HttpHandlerFuncDecorator) http.HandlerFunc {
    // 关键点4: 通过for循环遍历装饰器，完成对被装饰接口的装饰
    for _, decorator := range decorators {
        h = decorator(h)
    }

    ctx := context.Background()
    ctx, _ = context.WithCancel(ctx)
    ctx, _ = context.WithTimeout(ctx, time.Duration(1))
    ctx = context.WithValue(ctx, "key", "value")
    return h
}

// WithBasicAuth
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
    // 关键点7: 通过Decorate方法完成对hello对装饰
    http.HandleFunc("/hello", Decorate(hello, WithLogger, WithBasicAuth))
    // 启动http服务器
    http.ListenAndServe("localhost:8080", nil)
}

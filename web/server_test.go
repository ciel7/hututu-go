package web

import (
	"fmt"
	"sync"
	"testing"
)

func TestServer(t *testing.T) {
	//var h Server // New server
	//var h Server = &HTTPServer{}
	//h := &HTTPServer{}
	h := NewHTTPServer()
	// 注册路由
	//h.AddRoutes(http.MethodGet, "/user", func(ctx Context) {
	//	fmt.Println("处理第一件事")
	//}, func(ctx Context) {
	//	fmt.Println("处理第二件事")
	//})

	//handler1 := func(ctx *Context) {
	//	fmt.Println("处理第一件事")
	//}
	//handler2 := func(ctx *Context) {
	//	fmt.Println("处理第二件事")
	//}
	//h.AddRoutes(http.MethodGet, "/user", handler1, handler2)
	//h.AddRoute(http.MethodGet, "/user", func(ctx Context) {
	//	handler1(ctx)
	//	handler2(ctx)
	//})

	//h.addRoute(http.MethodGet, "/user", func(ctx *Context) {
	//	//handler1(ctx)
	//	//handler2(ctx)
	//	ctx.Resp.Write([]byte("Hello, Order Detail!"))
	//})

	h.Get("/order/abc", func(ctx *Context) {
		//handler1(ctx)
		//handler2(ctx)
		ctx.Resp.Write([]byte(fmt.Sprintf("Hello, %s", ctx.Req.URL.Path)))
	})

	h.Get("/login/:username", func(ctx *Context) {
		for k, v := range ctx.PathParams {
			ctx.Resp.Write([]byte(fmt.Sprintf("Hello, %s: %s", k, v)))
		}
	})

	//h.Get("/values/:id", func(ctx *Context) {
	//	val, err := ctx.PathValueV1("id").AsInt64()
	//	if err != nil {
	//		ctx.Resp.WriteHeader(400)
	//		ctx.Resp.Write([]byte(fmt.Sprintf("web: 缺少id")))
	//		return
	//	}
	//	ctx.Resp.Write([]byte(fmt.Sprintf("Hello, id  = %d", val)))
	//})

	h.Get("/values/:username", func(ctx *Context) {
		val := ctx.PathValueV1("username")
		if val.err != nil {
			ctx.RespJSON(400, "web: 缺少username")
			return
		}
		type User struct {
			Name string `json:"name"`
		}
		user := User{
			Name: val.val,
		}
		ctx.RespJSON(200, user)
	})

	h.Get("/user/:name", func(ctx *Context) {
		safeCtx := &SafeContext{
			Context: *ctx,
		}

		val := ctx.PathValueV1("name")
		if val.err != nil {
			err := safeCtx.RespJSON(404, "web: 缺少 name")
			if err != nil {
				safeCtx.RespJSON(404, err.Error())
			}
			return
		}
		type User struct {
			Name string `json:"name"`
		}
		user := User{
			Name: val.val,
		}
		safeCtx.RespJSON(200, user)
	})

	// 注意在使用 h.Get 之前需要把 var h Server = &HTTPServer{} 改成 h := &HTTPServer{}
	// 因为 Server 接口类型是没有 Get 方法的，Get是通过调用 AddRoute 实现的
	//h.Get("/user", handler1)

	// 用法一 完全委托给 http 包
	//http.ListenAndServe(":8701", h)
	//http.ListenAndServeTLS(":443", "", "", h)

	// 用法二 自己手动管
	h.Start(":8701")
}

type SafeContext struct {
	Context
	mutex sync.Mutex
}

func (c *SafeContext) RespJSONOK(val any) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.Context.RespJSONOK(val)
}

func (c *SafeContext) RespJSON(status int, val any) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.Context.RespJSON(status, val)
}

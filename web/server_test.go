package web

import (
	"fmt"
	"testing"
)

func TestServer(t *testing.T) {
	//var h Server // New server
	//var h Server = &HTTPServer{}
	h := &HTTPServer{}
	// 注册路由
	//h.AddRoutes(http.MethodGet, "/user", func(ctx Context) {
	//	fmt.Println("处理第一件事")
	//}, func(ctx Context) {
	//	fmt.Println("处理第二件事")
	//})

	handler1 := func(ctx Context) {
		fmt.Println("处理第一件事")
	}
	//handler2 := func(ctx Context) {
	//	fmt.Println("处理第二件事")
	//}
	//h.AddRoutes(http.MethodGet, "/user", handler1, handler2)
	//h.AddRoute(http.MethodGet, "/user", func(ctx Context) {
	//	handler1(ctx)
	//	handler2(ctx)
	//})

	// 注意在使用 h.Get 之前需要把 var h Server = &HTTPServer{} 改成 h := &HTTPServer{}
	// 因为 Server 接口类型是没有 Get 方法的，Get是通过调用 AddRoute 实现的
	h.Get("/user", handler1)

	// 用法一 完全委托给 http 包
	//http.ListenAndServe(":8701", h)
	//http.ListenAndServeTLS(":443", "", "", h)

	// 用法二 自己手动管
	h.Start(":8701")
}

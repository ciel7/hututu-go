package accesslog

import (
	"fmt"
	"hututu-go/web"
	"testing"
)

func TestMiddleware(t *testing.T) {
	// 第一步：创建"中间件构建器"，并设置自定义日志函数
	builder := MiddlewareBuilder{
		// 定义日志函数：收到log字符串后，打印到控制台
		logFunc: func(log string) {
			fmt.Println(log)
		},
	}
	// 第二步：用构建器生成最终的中间件（相当于"定制好一杯配料"）
	mdl := builder.Build()
	// 第三步：把中间件转换成"服务器配置选项"（相当于"告诉奶茶店：我要加这个配料"）
	opt := web.ServerWithMiddleware(mdl)
	// 第四步：创建HTTP服务器，传入配置选项（相当于"奶茶店组装奶茶：加了这个配料"）
	server := web.NewHTTPServer(opt)

	// 创建一个路由
	server.Get("/a/b/*", func(ctx *web.Context) {
		ctx.Resp.Write([]byte("Hello ,it's me."))
	})

	//req, err := http.NewRequest(http.MethodPost, "/a/b/c", nil)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//server.ServeHTTP(nil, req)
	// 第五步：启动服务器，监听 8701 端口（相当于"做好的奶茶端上桌，开始使用"）
	server.Start(":8701")
}

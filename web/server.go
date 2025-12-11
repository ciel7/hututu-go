package web

import (
	"fmt"
	"net"
	"net/http"
)

type HandleFunc func(ctx *Context)

// 如果 HTTPSServer 没有完全实现 Server 接口要求的所有方法，编译就会报错。
// 它是一种静态验证机制，确保类型满足接口契约，而不会在运行时才暴露问题。
// 这种方式在项目中被用来做隐式的接口实现检查。
// 确保一定会实现 Server 接口
var _ Server = &HTTPSServer{}

// Server 遇事不决先接口
// 不在这里实现Get、Post等，保证接口的小而美，可以在外面通过调用addRoute实现这些方法
type Server interface {
	http.Handler
	Start(string) error
	//Start1() error

	// addRoute 路由注册
	// method HTTP方法
	// path 路由
	// handleFunc 业务逻辑
	addRoute(method string, path string, handleFunc HandleFunc)

	// addRoutes 允许注册多个路由，没有必要提供，可以让用户自己去管
	// 如果允许注册多个，那么在实现的时候就要考虑，其中一个失败了，是否还允许继续执行下去；
	// 反过来，如果其中一个 HandleFunc 要中断执行，怎么中断。
	//addRoutes(method string, path string, handleFunc ...HandleFunc)
}

// HTTPSServer 变成了 HTTPServer 的装饰器
type HTTPSServer struct {
	HTTPServer
}

// HTTPServer HTTP服务器核心结构体
type HTTPServer struct {
	// Addr string // 可以改成这样，即创建的时候传递，而不是在 Start 的时候接收
	// *router
	// r *router
	router              // 路由实例（由newRouter()创建）
	mdls   []Middleware // 中间件列表

	log func(msg string, args ...any)
}

// HTTPServerOption 选项函数类型
// 本质是"接收*HTTPServer的函数"，用于修改服务器实例的属性
type HTTPServerOption func(server *HTTPServer)

// NewHTTPServerV1 这种方案不如 NewHTTPServer，缺乏扩展性
func NewHTTPServerV1(mdls ...Middleware) *HTTPServer {
	res := &HTTPServer{
		router: newRouter(),
		mdls:   mdls,
	}
	return res
}

// NewHTTPServer 创建 HTTP 服务器的构造函数
func NewHTTPServer(opts ...HTTPServerOption) *HTTPServer {
	// 创建默认的HTTPServer实例
	res := &HTTPServer{
		router: newRouter(), // 初始化路由（默认值）
		// mdls 没有显式初始化，默认是nil切片
		log: func(msg string, args ...any) {
			fmt.Printf(msg, args...)
		},
	}

	// 遍历所有传入的选项函数，逐个修改服务器实例
	for _, opt := range opts {
		opt(res) // 执行选项函数，把默认实例传进去修改
	}
	// 返回最终配置好的服务器实例
	return res
}

// ServerWithMiddleware 生成「设置中间件」的选项函数
// 「工厂函数」—— 接收用户想要设置的中间件列表；
// 返回一个「选项函数」（符合 HTTPServerOption 类型）；
// 把中间件列表赋值给 HTTPServer 的 mdls 字段
func ServerWithMiddleware(mdls ...Middleware) HTTPServerOption {
	// 返回一个符合HTTPServerOption类型的匿名函数
	return func(server *HTTPServer) {
		// 把传入的中间件列表赋值给服务器的mdls字段
		server.mdls = mdls
	}
}

//func NewHTTPServer() *HTTPServer {
//	return &HTTPServer{
//		router: newRouter(),
//	}
//}

// ServeHTTP 核心: 处理请求的入口
func (h *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// 在此实现框架代码: Context构建、路由匹配（之前需要注册路由）、执行业务逻辑
	ctx := &Context{
		Req:  request,
		Resp: writer,
	}
	// h.serve(ctx)

	// 最后一个是这个 var root func(ctx *Context) = h.serve
	root := h.serve

	// 利用最后一个 Middleware 不断往前回溯组装一个 中间件 chain
	// 从后往前，把后一个作为前一个执行顺序上的 next 1 2 3 4 5
	for i := len(h.mdls) - 1; i >= 0; i-- {
		root = h.mdls[i](root) // type Middleware func(next HandleFunc) HandleFunc
	}

	//root = serve
	//root = m3(root)  // root 现在是 m3(serve)
	//root = m2(root)  // root 现在是 m2(m3(serve))
	//root = m1(root)  // root 现在是 m1(m2(m3(serve)))
	//最终 root 是一个层层嵌套的调用链：m1 -> m2 -> m3 -> serve

	// 那么在这里执行的时候，就是从前往后执行了

	// 这里需要把 RespData 和 RespStatusCode 刷新到最终响应里面
	var m Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			// 就设置到了 RespData 和 RespStatusCode
			next(ctx)
			h.flashResp(ctx)
		}
	}

	//m(m1(m2(m3(serve))))
	root = m(root)
	root(ctx)

	// 展开后的执行顺序：
	//1. 调用 m 返回的函数，传入 ctx
	//2. 在 m 函数内部：调用 next(ctx) → 这个 next 是 m1(m2(m3(serve)))
	//3. 在 m1 函数内部：调用 next(ctx) → 这个 next 是 m2(m3(serve))
	//4. 在 m2 函数内部：调用 next(ctx) → 这个 next 是 m3(serve)
	//5. 在 m3 函数内部：调用 next(ctx) → 这个 next 是 serve
	//6. 在 serve 函数内部：执行业务逻辑（路由匹配、处理请求）
	//7. 返回到 m3，执行 m3 中 next 之后的代码
	//8. 返回到 m2，执行 m2 中 next 之后的代码
	//9. 返回到 m1，执行 m1 中 next 之后的代码
	//10. 返回到 flashMw，执行 flashResp 刷新响应
}

func (h *HTTPServer) flashResp(ctx *Context) {
	if ctx.RespStatusCode != 0 {
		ctx.Resp.WriteHeader(ctx.RespStatusCode)
	}
	n, err := ctx.Resp.Write(ctx.RespData)
	if err != nil || n != len(ctx.RespData) {
		h.log("http resp 写入异常: %v", err)
	}
}

func (h *HTTPServer) serve(ctx *Context) {
	// 接下来就是查找路由，并且执行命中的业务逻辑
	info, ok := h.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || info == nil || info.n == nil || info.n.handler == nil {
		// 路由没有命中，返回 404
		//ctx.Resp.WriteHeader(404)
		//_, _ = ctx.Resp.Write([]byte("NOT FOUND"))
		ctx.RespData = []byte("NOT FOUND")
		ctx.RespStatusCode = 404
		return
	}
	ctx.PathParams = info.pathParams
	ctx.MatchedRoute = info.n.route // 命中的路由
	info.n.handler(ctx)
}

func (h *HTTPServer) Start(addr string) error {
	// 也可以自己创建 Server
	// http.Server{}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("err", err.Error())
		return err
	}

	// Start 和 Start1的区别在于
	// 在这里 用户可以注册 after start 回调
	// 比如在 admin 注册一下这个实例；执行一些业务所需的前置条件
	return http.Serve(l, h)
}

//func (h *HTTPServer) Start1() error {
//	return http.ListenAndServe(h.Addr, h)
//}

//func (h *HTTPServer) addRoute(method string, path string, handleFunc HandleFunc) {
//	// 注册到路由树里
//	//panic("implement me")
//	fmt.Println("implement me addRoute")
//}

func (h *HTTPServer) Get(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodGet, path, handleFunc)
}

func (h *HTTPServer) Post(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodPost, path, handleFunc)
}

func (h *HTTPServer) Put(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodPut, path, handleFunc)
}

func (h *HTTPServer) Delete(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodDelete, path, handleFunc)
}

func (h *HTTPServer) Options(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodOptions, path, handleFunc)
}

//func (h *HTTPServer) addRoutes(method string, path string, handleFunc ...HandleFunc) {
//	panic("implement me")
//}

package web

import (
	"net"
	"net/http"
)

type HandleFunc func(ctx Context)

// 如果 HTTPSServer 没有完全实现 Server 接口要求的所有方法，编译就会报错。
// 它是一种静态验证机制，确保类型满足接口契约，而不会在运行时才暴露问题。
// 这种方式在项目中被用来做隐式的接口实现检查。
// 确保一定会实现 Server 接口
var _ Server = &HTTPSServer{}

// Server 遇事不决先接口
// 不在这里实现Get、Post等，保证接口的小而美，可以在外面通过调用AddRoute实现这些方法
type Server interface {
	http.Handler
	Start(string) error
	//Start1() error

	// AddRoute 路由注册
	// method HTTP方法
	// path 路由
	// handleFunc 业务逻辑
	AddRoute(method string, path string, handleFunc HandleFunc)

	// AddRoutes 允许注册多个路由，没有必要提供，可以让用户自己去管
	// 如果允许注册多个，那么在实现的时候就要考虑，其中一个失败了，是否还允许继续执行下去；
	// 反过来，如果其中一个 HandleFunc 要中断执行，怎么中断。
	AddRoutes(method string, path string, handleFunc ...HandleFunc)
}

// HTTPSServer 变成了 HTTPServer 的装饰器
type HTTPSServer struct {
	HTTPServer
}

type HTTPServer struct {
	Addr string // 可以改成这样，即创建的时候传递，而不是在 Start 的时候接收
}

// ServeHTTP 核心: 处理请求的入口
func (h *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// 在此实现框架代码: Context构建、路由匹配（之前需要注册路由）、执行业务逻辑
	ctx := &Context{
		Req:  request,
		Resp: writer,
	}

	h.serve(ctx)
}

func (h *HTTPServer) serve(ctx *Context) {
	// 接下来就是查找路由，并且执行命中的业务逻辑

}

func (h *HTTPServer) Start(addr string) error {
	// 也可以自己创建 Server
	// http.Server{}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// Start 和 Start1的区别在于
	// 在这里 用户可以注册 after start 回调
	// 比如在 admin 注册一下这个实例；执行一些业务所需的前置条件
	return http.Serve(l, h)
}

func (h *HTTPServer) Start1() error {
	return http.ListenAndServe(h.Addr, h)
}

func (h *HTTPServer) AddRoute(method string, path string, handleFunc HandleFunc) {
	// 注册到路由树里
	panic("implement me")
}

func (h *HTTPServer) Get(path string, handleFunc HandleFunc) {
	h.AddRoute(http.MethodGet, path, handleFunc)
}

func (h *HTTPServer) Post(path string, handleFunc HandleFunc) {
	h.AddRoute(http.MethodPost, path, handleFunc)
}

func (h *HTTPServer) Put(path string, handleFunc HandleFunc) {
	h.AddRoute(http.MethodPut, path, handleFunc)
}

func (h *HTTPServer) Delete(path string, handleFunc HandleFunc) {
	h.AddRoute(http.MethodDelete, path, handleFunc)
}

func (h *HTTPServer) Options(path string, handleFunc HandleFunc) {
	h.AddRoute(http.MethodOptions, path, handleFunc)
}

func (h *HTTPServer) AddRoutes(method string, path string, handleFunc ...HandleFunc) {
	panic("implement me")
}

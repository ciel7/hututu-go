package web

// Middleware 函数式的责任链模式 函数式的洋葱模式
type Middleware func(next HandleFunc) HandleFunc

// MiddlewareV1
type MiddlewareV1 interface {
	Invoke(next HandleFunc) HandleFunc
	//func(next HandleFunc) HandleFunc
}

// Interceptor 拦截器
type Interceptor interface {
	Before(ctx *Context)
	After(ctx *Context)
	Surround(ctx *Context)
}

type Chain []HandleFunc

type ChainV1 struct {
	handlers []HandleFunc
}

func (c *ChainV1) Run(ctx *Context) {
	for _, handler := range c.handlers {
		handler(ctx)
	}
}

type MiddlewareV2 func(ctx *Context) bool

type ChainV2 struct {
	handlers []MiddlewareV2
}

func (c *ChainV2) Run(ctx *Context) {
	for _, handler := range c.handlers {
		next := handler(ctx)
		if !next { // 中断执行
			return
		}
	}
}

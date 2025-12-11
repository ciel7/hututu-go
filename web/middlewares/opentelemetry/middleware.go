// package opentelemetry 为 HTTP 服务的每个请求创建一个 span，记录请求的关键信息，并支持跨进程的链路追踪
package opentelemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"hututu-go/web"
)

// Trace（追踪）	描述一个请求从入口到出口的完整链路（比如：用户请求 → 网关 → 服务 A → 服务 B）
// Span（跨度）	Trace 的最小单元，代表链路中的一个「操作」（比如：处理 HTTP 请求、调用数据库）
// Tracer	用来创建 Span 的「工具」，相当于 Span 的工厂
// 传播（Propagation）	跨进程传递 Trace 上下文（比如：服务 A 调用服务 B 时，把 Trace ID 传给 B）
// 属性（Attribute）	给 Span 附加的键值对元数据（比如：HTTP 方法、请求路径、状态码）
// TracerProvider	管理 Tracer 的「容器」，OTel 全局默认有一个，也可以自定义

// 标识这个追踪器的「归属」（OTel 规范要求，用于区分不同模块的 Tracer）
const instrumentationName = "github.com/ciel7/hututu-go/web/middlewares/opentelemetry"

// MiddlewareBuilder 构建中间件的结构体，核心字段是 Trace（Span 工厂）
type MiddlewareBuilder struct {
	//trace trace.Tracer
	Trace trace.Tracer
}

// NewMiddlewareBuilder 当 MiddlewareBuilder 定义的 trace 是私有时，可以参考
//func NewMiddlewareBuilder(trace trace.Tracer) *MiddlewareBuilder {
//	return &MiddlewareBuilder{
//		trace: trace,
//	}
//}

func (m MiddlewareBuilder) Build() web.Middleware {
	// 若用户没有传入自定的 Trace，就用 Otel 全局默认的 TracerProvider 创建一个
	if m.Trace == nil {
		m.Trace = otel.GetTracerProvider().Tracer(instrumentationName)
	}

	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			// ctx.Req.Context() 这种形式是只在进程内传递，我们还需要考虑跨进程的情况
			reqCtx := ctx.Req.Context()

			//从 HTTP 请求头中提取 Trace 上下文（比如 Trace ID、Parent Span ID）
			//比如用户请求「网关 → 你的服务」，网关已经创建了一个 Span，你的服务需要继承这个 Span，形成完整的链路；
			//TextMapPropagator：OTel 内置的传播器，默认从 HTTP Header 中读取 traceparent、tracestate 等字段；
			//HeaderCarrier：把 HTTP 请求头包装成 OTel 能读取的格式（适配层）。

			reqCtx = otel.GetTextMapPropagator().Extract(reqCtx, propagation.HeaderCarrier(ctx.Req.Header))

			// 创建一个新的 Span
			// 尝试和客户端的span结合在一起
			reqCtx, span := m.Trace.Start(reqCtx, "unknown")

			// 标记 Span 结束，OTel 会在此时采集 Span 的耗时、属性等数据
			// 如果不调用 End()，Span会被认为是「未完成」的，不会被上报，追踪数据丢失
			// defer 保证即使中间发生 panic，End()也会被执行（可结合recover，避免panic导致服务崩溃）
			//defer span.End()

			defer func() {
				// 只有执行完 next 才可能有值
				// 因为路由匹配通常是在业务处理函数中完成的
				// ctx.MatchedRoute 只有执行完 next 才会有值
				if ctx.MatchedRoute != "" {
					// 使用命中的路由作为Span名字
					span.SetName(ctx.MatchedRoute)
				}

				// 服务响应状态码
				span.SetAttributes(
					attribute.Int("http.status", ctx.RespStatusCode),
				)

				span.End()
			}()

			// 给 Span 附加键值对属性，会被采集到 OTel 后端
			span.SetAttributes(
				attribute.String("http.method", ctx.Req.Method),
				attribute.String("http.url", ctx.Req.URL.Path),
				attribute.String("http.scheme", ctx.Req.URL.Scheme),
				attribute.String("http.host", ctx.Req.Host),
				attribute.String("http.remote_addr", ctx.Req.RemoteAddr),
			)

			// 更新请求上下文
			// ctx.Req.WithContext 会复制一份 ctx，性能不是很好，能不这么做的时候尽量不做，这里权宜之计，只能这么做
			// 把包含 Span 的上下文设置回请求
			// 这样后续的处理函数就能拿到这个上下文
			// 它们创建的子 Span 会自动成为这个 Span 的子节点
			ctx.Req = ctx.Req.WithContext(reqCtx)

			// 执行下一个中间件 / 处理函数
			// 中间件的链式调用，先执行当前追踪逻辑，再把控制权交给下一个处理函数
			next(ctx)
		}
	}
}

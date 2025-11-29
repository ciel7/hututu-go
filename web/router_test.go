package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func TestRouter_addRoute(t *testing.T) {
	// 1. 构造路由树
	// 2. 验证路由树
	testRoutes := []struct {
		method string
		path   string
		//HandleFunc
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
	}

	// 新增路由树
	var mockHandler HandleFunc = func(ctx Context) {
	}

	r := newRouter()
	for _, route := range testRoutes {
		r.addRoute(route.method, route.path, mockHandler)
	}

	// 预期路由树
	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: &node{
				path:    "/",
				handler: mockHandler,
				children: map[string]*node{
					"user": &node{
						path:    "user",
						handler: mockHandler,
						children: map[string]*node{
							"home": &node{
								path:    "home",
								handler: mockHandler,
							},
						},
					},
					"order": &node{
						path: "order",
						//handler: mockHandler,
						children: map[string]*node{
							"detail": &node{
								path:    "detail",
								handler: mockHandler,
							},
						},
					},
				},
			},
			http.MethodPost: &node{
				path: "/",
				//handler: mockHandler,
				children: map[string]*node{
					"order": &node{
						path: "order",
						//handler: mockHandler,
						children: map[string]*node{
							"create": &node{
								path:    "create",
								handler: mockHandler,
							},
						},
					},
					"login": &node{
						path:    "login",
						handler: mockHandler,
					},
				},
			},
		},
	}

	// 断言两棵树一致: 新增路由树 和 预期路由树
	// assert.Equal(t, r, wantRouter) // 该方法不可行，因为HandleFunc是不可比的
	msg, ok := wantRouter.equal(r)
	assert.True(t, ok, msg)

	r = newRouter()
	// 验证在特定条件下，一个函数是否会触发程序崩溃（Panic）
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	}, "web: 路径不可为空")

	r = newRouter()
	// 验证在特定条件下，一个函数是否会触发程序崩溃（Panic）
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "user", mockHandler)
	}, "web: 路径必须以 / 开头")

	r = newRouter()
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/user/", mockHandler)
	}, "web: 路径不能以 / 结尾")

	r = newRouter()
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/user//home", mockHandler)
	}, "web: 不能有连续的 /")

	r = newRouter()
	r.addRoute(http.MethodGet, "/", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	}, "web: 路由冲突 重复注册[/]")

	r = newRouter()
	r.addRoute(http.MethodGet, "/a/b/c/d", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/b/c/d", mockHandler)
	}, "web: 路由冲突 重复注册[/a/b/c/d]")

	// 可用的 HTTP method 要不要校验 ---> 不需要, 把 AddRoute 改成私有的，让用户无法调用，那么用户只能使用框架提供的 Get、Post之类的方法
	// r.AddRoute("aaa", "/a/b/c/d", mockHandler)
	// r.addRoute("aaa", "/a/b/c/d", mockHandler)
	// mockHandler 为 nil 呢？要不要校验
}

func (r *router) equal(y *router) (string, bool) {
	for method, node := range r.trees {
		// 比较 method
		dst, ok := y.trees[method]
		if !ok {
			return fmt.Sprintf("找不到对应的 HTTP method"), false
		}

		// node 和 dst 应该一致
		msg, ok := node.equal(dst)
		if !ok {
			return msg, false
		}
	}

	return "", true
}

func (n *node) equal(y *node) (string, bool) {
	// 比较 path
	if n.path != y.path {
		return fmt.Sprintf("节点路径不匹配"), false
	}

	if len(n.children) != len(y.children) {
		return fmt.Sprintf("子节点数量不相等"), false
	}

	// 比较 handlefunc
	nHandler := reflect.ValueOf(n.handler)
	yHandler := reflect.ValueOf(y.handler)
	if nHandler != yHandler {
		return fmt.Sprintf("handler 不相等"), false
	}

	for path, node := range n.children {
		dst, ok := y.children[path]
		if !ok {
			return fmt.Sprintf("子节点 %s 不存在", path), false
		}
		msg, ok := node.equal(dst)
		if !ok {
			return msg, false
		}
	}

	return "", true
}

func TestRouter_findRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
		//HandleFunc
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
	}

	r := newRouter()
	// 新增路由树
	var mockHandler HandleFunc = func(ctx Context) {
	}

	for _, route := range testRoutes {
		// 注册路由
		r.addRoute(route.method, route.path, mockHandler)
	}

	testCases := []struct {
		name   string
		method string
		path   string

		wantFound bool
		wantNode  *node
	}{
		{
			name:      "method not found",
			method:    http.MethodOptions,
			path:      "/order/detail",
			wantFound: false,
		},
		{
			name:      "order detail",
			method:    http.MethodGet,
			path:      "/order/detail",
			wantFound: true,
			wantNode: &node{
				handler: mockHandler,
				path:    "detail",
			},
		},
		{
			name:      "order",
			method:    http.MethodGet,
			path:      "/order",
			wantFound: true,
			wantNode: &node{
				path: "order",
				children: map[string]*node{
					"detail": &node{
						path:    "detail",
						handler: mockHandler,
					},
				},
			},
		},
		{
			name:      "path not found",
			method:    http.MethodDelete,
			path:      "/",
			wantFound: false,
		},
		{
			name:      "root",
			method:    http.MethodGet,
			path:      "/",
			wantFound: true,
			wantNode: &node{
				path:    "/",
				handler: mockHandler,
				children: map[string]*node{
					"user": &node{
						path:    "user",
						handler: mockHandler,
						children: map[string]*node{
							"home": &node{
								path:    "home",
								handler: mockHandler,
							},
						},
					},
					"order": &node{
						path: "order",
						children: map[string]*node{
							"detail": &node{
								path:    "detail",
								handler: mockHandler,
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		// t.Run 会创建一个新的子测试，
		// 第一个参数 tc.name 是子测试的名称，通常具有描述性，比如 "found_exact_match" 或 "not_found_with_wildcard".
		// 第二个参数是一个匿名函数，它是子测试的主体。注意：这个函数内部的 t 是一个新的 *testing.T 实例，专门用于这个子测试。
		t.Run(tc.name, func(t *testing.T) {
			findNode, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.wantFound, found)
			if !found {
				return
			}
			//assert.Equal(t, tc.wantNode.path, findNode.path)
			//assert.Equal(t, tc.wantNode.children, findNode.children)
			msg, ok := tc.wantNode.equal(findNode)
			assert.True(t, ok, msg)
		})
	}
}

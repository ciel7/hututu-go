package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func TestRouter_AddRoute(t *testing.T) {
	// 1. 构造路由树
	// 2. 验证路由树
	testRoutes := []struct {
		method string
		path   string
		//HandleFunc
	}{
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
	}

	// 新增路由树
	var mockHandler HandleFunc = func(ctx Context) {
	}

	r := newRouter()

	for _, route := range testRoutes {
		r.AddRoute(route.method, route.path, mockHandler)
	}

	// 预期路由树
	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: &node{
				path: "/",
				children: map[string]*node{
					"user": &node{
						path: "user",
						children: map[string]*node{
							"home": &node{
								path:    "home",
								handler: mockHandler,
							},
						},
					},
				},
			},
		},
	}

	// 断言两棵树一致: 新增路由树 和 预期路由树
	// assert.Equal(t, r, wantRouter) // 该方法不可行，因为HandleFunc是不可比的
	msg, ok := wantRouter.equal(r)
	assert.True(t, ok, msg)
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

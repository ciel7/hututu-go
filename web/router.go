package web

import "strings"

// 代表路由树(森林)
// 支持对路由树的操作
type router struct {
	// Beego Gin HTTP method 对应一棵树
	// GET 有一棵树，POST也有一棵树，...

	// http method => 路由树根节点
	trees map[string]*node
}

func newRouter() *router {
	return &router{
		trees: map[string]*node{},
	}
}

func (r *router) AddRoute(method string, path string, handleFunc HandleFunc) {
	// 首先需要找到树，需要知道将现在的节点加到哪里
	// 那就需要从根找
	root, ok := r.trees[method]
	if !ok {
		// 如果不存在该 HTTP 的 method，那么就需要新增一个根节点
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}

	// 切割 path
	path = strings.TrimLeft(path, "/")
	// 目的是要把path的每一段都加入到路由中去
	// 如果哪段没有，就要创建哪段
	segs := strings.Split(path, "/")
	for _, seg := range segs {
		// 递归下去，找准位置
		// 如果中途有节点不存在，就要创建该节点
		children := root.childOrCreate(seg)
		root = children // 下次从 children 继续找
	}

	root.handler = handleFunc
}

func (n *node) childOrCreate(seg string) *node {
	if n.children == nil {
		n.children = map[string]*node{}
	}

	res, ok := n.children[seg]
	if !ok {
		res = &node{
			path: seg,
		}
		n.children[seg] = res
	}
	return res
}

type node struct {
	path string

	// 子 path 到子节点的映射
	children map[string]*node

	// 用户注册的业务逻辑
	handler HandleFunc
}

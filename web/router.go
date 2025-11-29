package web

import (
	"fmt"
	"strings"
)

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

// AddRoute 需要加一些限制
// path 必须以 / 开头，不能以 / 结尾，中间也不可以有连续的 //
func (r *router) addRoute(method string, path string, handleFunc HandleFunc) {
	if path == "" {
		panic("web: 路径不可为空")
	}
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

	if path[0] != '/' {
		panic("web: 路径必须以 / 开头")
	}

	if path != "/" && path[len(path)-1] == '/' {
		panic("web: 路径不能以 / 结尾")
	}

	// 中间连续 //，可以用 strings.contains("//") 检测
	// 也可以在下面的 遍历 segs 去做

	// 如果注册根节点需要特殊处理
	// 因为后续直接用strings.Split会导致segs元素是空字符串
	if path == "/" {
		// 根节点重复注册
		if root.handler != nil {
			panic("web: 路由冲突 重复注册[/]")
		}
		root.handler = handleFunc
		return
	}

	// 切割 path
	path = strings.TrimLeft(path, "/")
	// 目的是要把path的每一段都加入到路由中去
	// 如果哪段没有，就要创建哪段
	segs := strings.Split(path, "/")
	for _, seg := range segs {
		if seg == "" {
			panic("web: 不能有连续的 /")
		}
		// 递归下去，找准位置
		// 如果中途有节点不存在，就要创建该节点
		children := root.childOrCreate(seg)
		root = children // 下次从 children 继续找
	}

	if root.handler != nil {
		panic(fmt.Sprintf("web: 路由冲突，重复注册[%s]", root.path))
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

package web

import (
	"fmt"
	"regexp"
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

func newRouter() router {
	return router{
		trees: map[string]*node{},
	}
}

// addRoute 需要加一些限制
// path 必须以 / 开头，不能以 / 结尾，中间也不可以有连续的 //
func (r *router) addRoute(method string, path string, handleFunc HandleFunc) {
	if path == "" {
		panic("web: 路由是空字符串")
	}
	if path[0] != '/' {
		panic("web: 路由必须以 / 开头")
	}
	if path != "/" && path[len(path)-1] == '/' {
		panic("web: 路由不能以 / 结尾")
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

	// 中间连续 //，可以用 strings.contains("//") 检测
	// 也可以在下面的 遍历 segs 去做

	// 如果注册根节点需要特殊处理
	// 因为后续直接用strings.Split会导致segs元素是空字符串
	if path == "/" {
		// 根节点重复注册
		if root.handler != nil {
			panic("web: 路由冲突[/]")
		}
		root.handler = handleFunc
		return
	}

	// 切割 path
	// 目的是要把path的每一段都加入到路由中去
	// 如果哪段没有，就要创建哪段
	segs := strings.Split(path[1:], "/")
	for _, seg := range segs {
		if seg == "" {
			panic(fmt.Sprintf("web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [%s]", path))
		}
		// 递归下去，找准位置
		// 如果中途有节点不存在，就要创建该节点
		//child := root.childOrCreate(seg)
		//root = child // 下次从 children 继续找
		root = root.childOrCreate(seg)
	}

	if root.handler != nil {
		panic(fmt.Sprintf("web: 路由冲突[%s]", path))
	}
	root.handler = handleFunc
}

func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	// 沿着树深度查下去
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}

	if path == "/" {
		return &matchInfo{n: root}, true
	}

	if path == "" {
		return &matchInfo{
			n: root,
		}, true
	}

	// 把前置后置的 / 都去掉
	path = strings.Trim(path, "/")
	segs := strings.Split(path, "/")
	mi := &matchInfo{}
	for _, seg := range segs {
		var child *node
		child, ok = root.childOf(seg)
		if !ok {
			// 最后一段 *
			if root.typ == nodeTypeAny {
				mi.n = root
				return mi, true
			}
			return nil, false
		}

		if child.paramName != "" {
			mi.addValue(child.paramName, seg)
		}

		root = child
	}
	// 代表确实有这个节点
	// 但该节点是不是用户注册有handler的，就不一定了
	// 下面的则表示确实有这个节点 && 该节点有注册的handler
	//return root, root.handler != nil

	mi.n = root
	return mi, true
}

func (n *node) childOrCreate(path string) *node {
	if path == "*" {
		if n.paramChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [%s]", path))
		}
		if n.regChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有正则路由。不允许同时注册通配符路由和正则路由 [%s]", path))
		}
		if n.starChild == nil {
			n.starChild = &node{path: path, typ: nodeTypeAny}
		}
		return n.starChild
	}

	// 以 : 开头，需要进一步解析，判断是参数路由还是正则路由
	if path[0] == ':' {
		paramName, expr, isReg := n.parseParam(path)
		if isReg {
			return n.childOrCreateReg(path, expr, paramName)
		}
		return n.childOrCreateParam(path, paramName)
	}

	if n.children == nil {
		n.children = make(map[string]*node)
	}
	child, ok := n.children[path]
	if !ok {
		child = &node{path: path, typ: nodeTypeStatic}
		n.children[path] = child
	}
	return child
}

func (n *node) childOrCreateReg(path, expr, paramName string) *node {
	if n.starChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [%s]", path))
	}
	if n.paramChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册正则路由和参数路由 [%s]", path))
	}
	if n.regChild != nil {
		// 同一段注册了两个不一样的路由时，如 :id(xxx) :name(xxx)
		if n.regChild.regExpr.String() != expr || n.paramName != paramName {
			panic(fmt.Sprintf("web: 路由冲突，正则路由冲突，已有 %s，新注册 %s", n.regChild.regExpr.String(), expr))
		}
	} else {
		regExpr, err := regexp.Compile(expr)
		if err != nil {
			panic(fmt.Errorf("web: 正则表达式错误 %w", err))
		}
		n.regChild = &node{
			path:      path,
			paramName: paramName,
			regExpr:   regExpr,
			typ:       nodeTypeReg,
		}
	}
	return n.regChild
}
func (n *node) childOrCreateParam(path, paramName string) *node {
	if n.regChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有正则路由。不允许同时注册正则路由和参数路由 [%s]", path))
	}
	if n.starChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [%s]", path))
	}

	if n.paramChild != nil {
		if n.paramChild.path != path {
			panic(fmt.Sprintf("web: 路由冲突，参数路由冲突，已有 :%s，新注册 :%s", n.paramChild.paramName, paramName))
		}
	} else {
		n.paramChild = &node{
			path:      path,
			paramName: paramName,
			typ:       nodeTypeParam,
		}
	}
	return n.paramChild
}

// childOf 匹配优先级 静态匹配 > 路径参数匹配 > 通配符匹配
// 参数
// *node 命中的子节点
// bool 节点是否被命中
func (n *node) childOf(path string) (*node, bool) {
	if n.children == nil {
		return n.childOfNonStatic(path)
	}

	child, ok := n.children[path]
	if !ok {
		return n.childOfNonStatic(path)
	}
	return child, ok
}

// childOfNonStatic 从非静态匹配的子节点中按照优先级查找
func (n *node) childOfNonStatic(path string) (*node, bool) {
	if n.regChild != nil {
		if n.regChild.regExpr.Match([]byte(path)) {
			return n.regChild, true
		}
	}
	if n.paramChild != nil {
		return n.paramChild, true
	}
	return n.starChild, n.starChild != nil
}

// parseParam
// 返回值
// 1. 参数名字
// 2. 正则表达式
// 3. 是否是正则路由: true 是
//func (n *node) parseParam(path string) (string, string, bool) {
//	// 去除 :
//	path = path[1:]
//	// 要求正则表达式格式 :paramName(xxx)
//	segs := strings.SplitN(path, "(", 2)
//	if len(segs) == 2 {
//		expr := segs[1]
//		if strings.HasSuffix(expr, ")") {
//			return segs[0], expr[:len(expr)-1], true
//		}
//	}
//	return path, "", false
//}

// 编译正则表达式（全局变量，仅编译一次，提高性能）
var paramRegex = regexp.MustCompile(`^([a-zA-Z0-9_]+)\((.*)\)$`)

func (n *node) parseParam(path string) (string, string, bool) {
	// 1. 先判断路径是否以 : 开头，再去除 :（保持原有逻辑的第一步）
	if len(path) == 0 || path[0] != ':' {
		return "", "", false // 不是参数格式，直接返回失败
	}
	path = path[1:] // 去除 : 后的路径（如 "paramName(xxx)"）

	// 2. 用正则表达式匹配路径（替代 SplitN 和 HasSuffix）
	// 正则规则：^([a-zA-Z0-9_]+)\((.*)\)$
	// - ^：匹配字符串开头
	// - ([a-zA-Z0-9_]+)：捕获参数名（仅允许字母、数字、下划线，与原有逻辑兼容）
	// - \((.*)\)：捕获括号内的表达式（支持任意字符，包括嵌套括号）
	// - $：匹配字符串结尾（确保整个路径完全符合格式）
	matches := paramRegex.FindStringSubmatch(path)

	// 3. 验证匹配结果（长度为 3 表示完全匹配：整个字符串 + 2 个捕获组）
	if len(matches) == 3 {
		paramName := matches[1] // 第一个捕获组：参数名（如 "paramName"）
		expr := matches[2]      // 第二个捕获组：括号内的表达式（如 "xxx"）
		return paramName, expr, true
	}

	// 4. 若正则匹配失败，退化为原有逻辑：返回整个路径作为参数名，无表达式
	return path, "", false
}

type nodeType int

const (
	// 静态路由
	nodeTypeStatic = iota

	// 正则路由
	nodeTypeReg

	// 路径参数路由
	nodeTypeParam

	// 通配符路由
	nodeTypeAny
)

// node 代表路由树的节点
// 路由树的匹配顺序是:
// 1. 静态路由完全匹配
// 2. 正则匹配，形式：param_name(reg_expr)
// 3. 路径参数匹配，形式：param_name
// 4. 通配符匹配，形式：*
// 不支持回溯匹配
type node struct {
	typ nodeType

	path string

	// 静态匹配节点
	// 子 path 到子节点的映射
	children map[string]*node

	// 通配符 * 表达的节点，任意匹配
	starChild *node

	// 路径参数
	paramChild *node
	// 参数名：正则路由和参数路由都会使用该字段
	paramName string

	// 正则表达式
	regChild *node
	regExpr  *regexp.Regexp

	// 用户注册的业务逻辑
	handler HandleFunc
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
}

func (m *matchInfo) addValue(key string, value string) {
	if m.pathParams == nil {
		// 大多数情况，参数路径只会有一段
		m.pathParams = map[string]string{key: value}
	}
	m.pathParams[key] = value
}

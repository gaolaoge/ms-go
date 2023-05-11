package ms_go

import (
	"fmt"
	"net/http"
)

/*
1. 静态路由，
2. 支持 分组，
3. 支持 restful ，
4. 动态路由，尚未支持
*/

type HandleFunc func(ctx *Context)
type MethodType = string

const (
	ANY  MethodType = "ANY"
	GET  MethodType = "GET"
	POST MethodType = "POST"
)

type MethodHandleMap map[MethodType]HandleFunc
type HandleFuncMap map[string]MethodHandleMap

type routerGroup struct {
	name          string
	handleFuncMap HandleFuncMap
	treeNode      *treeNode
}

func (rg *routerGroup) Add(name string, handleFunc HandleFunc) {
	initRouterGroup(rg, rg.name+name)
	rg.handleFuncMap[rg.name+name][ANY] = handleFunc
}

func (rg *routerGroup) Get(name string, handleFunc HandleFunc) {
	initRouterGroup(rg, rg.name+name)
	rg.handleFuncMap[rg.name+name][GET] = handleFunc
}

func (rg *routerGroup) Post(name string, handleFunc HandleFunc) {
	initRouterGroup(rg, rg.name+name)
	rg.handleFuncMap[rg.name+name][POST] = handleFunc
}

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	// TODO 这里需要判断 group 是否已被创建过
	newGroup := &routerGroup{
		name:          name,
		handleFuncMap: make(HandleFuncMap),
		treeNode:      &treeNode{name: name, children: make([]*treeNode, 0)},
	}
	r.routerGroups = append(r.routerGroups, newGroup)
	return newGroup
}

func (r *router) Add(name string, handleFunc HandleFunc) {
	initRouterGroup(r.routerGroups[0], name)
	r.routerGroups[0].handleFuncMap[name][ANY] = handleFunc
}

func (r *router) Get(name string, handleFunc HandleFunc) {
	initRouterGroup(r.routerGroups[0], name)
	r.routerGroups[0].handleFuncMap[name][GET] = handleFunc
}

func (r *router) Post(name string, handleFunc HandleFunc) {
	initRouterGroup(r.routerGroups[0], name)
	r.routerGroups[0].handleFuncMap[name][POST] = handleFunc
}

type Engine struct {
	router
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. 遍历 groups 匹配 treeNode
	// 2. 在 methodMap 中匹配 method
	curMethod := r.Method
	for _, group := range e.routerGroups {
		node := group.treeNode.Get(r.RequestURI)
		if node != nil {
			resultMap := group.handleFuncMap[node.routerName]
			handle := resultMap[ANY]
			if handle != nil {
				handle(&Context{w, r})
				return
			}
			handle = resultMap[curMethod]
			if handle != nil {
				handle(&Context{w, r})
				return
			}
		}
	}
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `%s %s is not found \n`, curMethod, r.RequestURI)
}

func (e *Engine) Run() {
	http.Handle("/", e)
	http.ListenAndServe(":8081", nil)
}

func New() *Engine {
	return &Engine{
		router{
			routerGroups: []*routerGroup{{
				name:          "default",
				handleFuncMap: make(HandleFuncMap),
				treeNode:      &treeNode{name: "/", children: make([]*treeNode, 0)},
			}},
		},
	}
}

func initRouterGroup(root *routerGroup, name string) {
	if root.handleFuncMap[name] == nil {
		root.handleFuncMap[name] = make(MethodHandleMap, 0)
	}
	root.treeNode.Put(name)
}

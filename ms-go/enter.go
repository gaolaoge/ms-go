package ms_go

import (
	"fmt"
	"net/http"
)

type HandleFunc func(ctx *Context)
type HandleFuncMap map[string]HandleFunc
type MethodType = string

const (
	ANY  MethodType = "ANY"
	GET  MethodType = "GET"
	POST MethodType = "POST"
)

type MethodHandleMap map[MethodType]HandleFuncMap

type routerGroup struct {
	name                string
	handleMethodFuncMap MethodHandleMap
}

func (rg *routerGroup) Add(name string, handleFunc HandleFunc) {
	initRouterGroup(rg, ANY)
	rg.handleMethodFuncMap[ANY][name] = handleFunc
}

func (rg *routerGroup) Get(name string, handleFunc HandleFunc) {
	initRouterGroup(rg, GET)
	rg.handleMethodFuncMap[GET][name] = handleFunc
}

func (rg *routerGroup) Post(name string, handleFunc HandleFunc) {
	initRouterGroup(rg, POST)
	rg.handleMethodFuncMap[POST][name] = handleFunc
}

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	// TODO 这里需要判断 group 是否已被创建过
	newGroup := &routerGroup{
		name:                name,
		handleMethodFuncMap: make(MethodHandleMap),
	}
	r.routerGroups = append(r.routerGroups, newGroup)
	return newGroup
}

func (r *router) Add(name string, handleFunc HandleFunc) {
	initRouterGroup(r.routerGroups[0], ANY)
	r.routerGroups[0].handleMethodFuncMap[ANY][name] = handleFunc
}

func (r *router) Get(name string, handleFunc HandleFunc) {
	initRouterGroup(r.routerGroups[0], GET)
	r.routerGroups[0].handleMethodFuncMap[GET][name] = handleFunc
}

func (r *router) Post(name string, handleFunc HandleFunc) {
	initRouterGroup(r.routerGroups[0], POST)
	r.routerGroups[0].handleMethodFuncMap[POST][name] = handleFunc
}

type Engine struct {
	router
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. 遍历 groups 匹配到 URL
	// 2. 在 methodMap 中匹配 method
	curMethod := r.Method
	var url string
	for _, group := range e.routerGroups {

		if len(group.handleMethodFuncMap[ANY]) > 0 {
			for name, handle := range group.handleMethodFuncMap[ANY] {
				if group.name == "default" {
					url = name
				} else {
					url = group.name + name
				}
				if url == r.RequestURI {
					handle(&Context{w, r})
					return
				}
			}
		}
		if len(group.handleMethodFuncMap[curMethod]) > 0 {
			for name, handle := range group.handleMethodFuncMap[curMethod] {
				if group.name == "default" {
					url = name
				} else {
					url = group.name + name
				}
				if url == r.RequestURI {
					handle(&Context{w, r})
					return
				}
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
				name:                "default",
				handleMethodFuncMap: map[MethodType]HandleFuncMap{ANY: make(HandleFuncMap)},
			}},
		},
	}
}

func initRouterGroup(root *routerGroup, method MethodType) {
	if len(root.handleMethodFuncMap[method]) == 0 {
		root.handleMethodFuncMap[method] = map[string]HandleFunc{}
	}
}

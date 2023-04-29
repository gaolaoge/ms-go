package ms_go

import (
	"fmt"
	"net/http"
)

type HandleFunc func(w http.ResponseWriter, r *http.Request)

type routerGroup struct {
	name             string
	handleFuncMap    map[string]HandleFunc
	handlerMethodMap map[string][]string
}

func (rg *routerGroup) Add(name string, handleFunc HandleFunc) {
	rg.handleFuncMap[name] = handleFunc
	rg.handlerMethodMap["Any"] = append(rg.handlerMethodMap["Any"], name)
}

func (rg *routerGroup) Get(name string, handleFunc HandleFunc) {
	rg.handleFuncMap[name] = handleFunc
	rg.handlerMethodMap[http.MethodGet] = append(rg.handlerMethodMap[http.MethodGet], name)
}

func (rg *routerGroup) Post(name string, handleFunc HandleFunc) {
	rg.handleFuncMap[name] = handleFunc
	rg.handlerMethodMap[http.MethodPost] = append(rg.handlerMethodMap[http.MethodPost], name)
}

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	// TODO 这里需要判断 group 是否已被创建过
	newGroup := &routerGroup{
		name:             name,
		handleFuncMap:    make(map[string]HandleFunc),
		handlerMethodMap: make(map[string][]string),
	}
	r.routerGroups = append(r.routerGroups, newGroup)
	return newGroup
}

func (r *router) Add(name string, handleFunc HandleFunc) {
	_default := r.routerGroups[0]
	_default.handleFuncMap[name] = handleFunc
	_default.handlerMethodMap["Any"] = append(_default.handlerMethodMap["Any"], name)
}

func (r *router) Get(name string, handleFunc HandleFunc) {
	_default := r.routerGroups[0]
	_default.handleFuncMap[name] = handleFunc
	_default.handlerMethodMap[http.MethodGet] = append(_default.handlerMethodMap[http.MethodGet], name)
}

func (r *router) Post(name string, handleFunc HandleFunc) {
	_default := r.routerGroups[0]
	_default.handleFuncMap[name] = handleFunc
	_default.handlerMethodMap[http.MethodPost] = append(_default.handlerMethodMap[http.MethodPost], name)
}

type Engine struct {
	router
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. 遍历 groups 匹配到 URL
	// 2. 在 methodMap 中匹配 method
	method := r.Method
	for _, group := range e.routerGroups {
		for name, handle := range group.handleFuncMap {
			var url string
			if group.name == "default" {
				url = name
			} else {
				url = group.name + name
			}
			if r.RequestURI == url {
				routers, ok := group.handlerMethodMap["Any"]
				if ok {
					for _, routerName := range routers {
						if routerName == name {
							handle(w, r)
							return
						}
					}
				}
				routers, ok = group.handlerMethodMap[method]
				if ok {
					for _, routerName := range routers {
						if routerName == name {
							handle(w, r)
							return
						}
					}
				}
				w.WriteHeader(http.StatusMethodNotAllowed)
				fmt.Fprintf(w, `%s %s not allowed \n`, r.RequestURI, method)
				return
			}
		}
	}
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `%s not found \n`, r.RequestURI)
}

func (e *Engine) Run() {
	http.Handle("/", e)
	http.ListenAndServe(":8081", nil)
}

func New() *Engine {
	return &Engine{
		router{
			routerGroups: []*routerGroup{{
				name:             "default",
				handleFuncMap:    make(map[string]HandleFunc),
				handlerMethodMap: make(map[string][]string)}},
		},
	}
}

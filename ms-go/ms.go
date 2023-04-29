package ms_go

import (
	"net/http"
)

type HandleFunc func(w http.ResponseWriter, r *http.Request)

type routerGroup struct {
	name          string
	handleFuncMap map[string]HandleFunc
}

func (rg *routerGroup) Add(name string, handleFunc HandleFunc) {
	rg.handleFuncMap[name] = handleFunc
}

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	// TODO 这里需要判断 group 是否已被创建过
	newGroup := &routerGroup{
		name:          name,
		handleFuncMap: make(map[string]HandleFunc),
	}
	r.routerGroups = append(r.routerGroups, newGroup)
	return newGroup
}

func (r *router) Add(name string, handleFunc HandleFunc) {
	r.routerGroups[0].handleFuncMap[name] = handleFunc
}

type Engine struct {
	router
}

func (e *Engine) Run() {
	// TODO 尚未添加对"/"的处理，可以在此处建立一些规则
	for _, group := range e.routerGroups {
		for key, value := range group.handleFuncMap {
			var path string
			if group.name == "default" {
				path = key
			} else {
				path = group.name + key
			}
			http.HandleFunc(path, value)
		}
	}
	http.ListenAndServe(":8081", nil)
}

func New() *Engine {
	return &Engine{
		router{
			routerGroups: []*routerGroup{{"default", make(map[string]HandleFunc)}},
		},
	}
}

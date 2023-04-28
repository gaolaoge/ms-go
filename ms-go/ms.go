package ms_go

import (
	"net/http"
)

type HandleFunc func(w http.ResponseWriter, r *http.Request)

type router struct {
	handleFuncMap map[string]HandleFunc
}

func (r *router) Add(name string, handleFunc HandleFunc) {
	r.handleFuncMap[name] = handleFunc
}

type Engine struct {
	router
}

func (e *Engine) Run() {
	for key, value := range e.handleFuncMap {
		http.HandleFunc(key, value)
	}

	http.ListenAndServe(":8081", nil)
}

func New() *Engine {
	return &Engine{
		router: router{handleFuncMap: make(map[string]HandleFunc)},
	}
}

package ms_go

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/gaolaoge/ms-go/render"
)

func execute(pkg *HandlePkg, ctx *Context, root *routerGroup) {
	handle := pkg.handleFunc
	soleMiddlewareFunc := pkg.middlewareFunc
	if len(soleMiddlewareFunc) > 0 {
		for _, middlewareFunc := range soleMiddlewareFunc {
			handle = middlewareFunc(handle)
		}
	}
	if len(root.middlewareFuncSlice) > 0 {
		for _, middlewareFunc := range root.middlewareFuncSlice {
			handle = middlewareFunc(handle)
		}
	}
	handle(ctx)
}

type Engine struct {
	router
	funcMap    template.FuncMap
	HTMLRender *render.HTMLRender
}

func (e *Engine) SetTemplate(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

func (e *Engine) LoadTemplate(pattern string) {
	template := template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
	e.HTMLRender.Template = template
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 步骤：
	// 1. 遍历 groups 匹配 treeNode
	// 2. 在 methodMap 中匹配 method
	ctx := &Context{w, r, e}
	for _, group := range e.routerGroups {
		node := group.treeNode.Get(r.RequestURI)
		if node != nil {
			resultMap := group.handleFuncMap[node.routerName]
			if _, ok := resultMap[ANY]; ok {
				execute(resultMap[ANY], ctx, group)
				return
			}
			if _, ok := resultMap[r.Method]; ok {
				execute(resultMap[r.Method], ctx, group)
				return
			}
		}
	}
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `%s %s is not found \n`, r.Method, r.RequestURI)
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
		nil, &render.HTMLRender{},
	}
}

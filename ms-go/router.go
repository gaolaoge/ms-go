package ms_go

/*
1. 静态路由，
2. 支持 分组，
3. 以上支持 restful ，
4. 动态路由，尚未支持
*/

type MethodType = string

const (
	ANY  MethodType = "ANY"
	GET  MethodType = "GET"
	POST MethodType = "POST"
)

type HandleFunc func(ctx *Context)
type HandlePkg struct {
	handleFunc     HandleFunc
	middlewareFunc []MiddlewareFunc
}
type MethodHandleMap map[MethodType]*HandlePkg
type HandleFuncMap map[string]MethodHandleMap

type routerGroup struct {
	name                string
	handleFuncMap       HandleFuncMap
	middlewareFuncSlice []MiddlewareFunc
	treeNode            *treeNode
}

func (rg *routerGroup) Add(name string, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	addToRouter(rg, rg.name+name, ANY, handleFunc, middlewareFunc...)
}

func (rg *routerGroup) Get(name string, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	addToRouter(rg, rg.name+name, GET, handleFunc, middlewareFunc...)
}

func (rg *routerGroup) Post(name string, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	addToRouter(rg, rg.name+name, POST, handleFunc, middlewareFunc...)
}

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	// TODO 这里需要判断 group 是否已被创建过
	newGroup := &routerGroup{
		name:                name,
		handleFuncMap:       make(HandleFuncMap),
		middlewareFuncSlice: make([]MiddlewareFunc, 0),
		treeNode:            &treeNode{name: name, children: make([]*treeNode, 0)},
	}
	r.routerGroups = append(r.routerGroups, newGroup)
	return newGroup
}

func (r *router) Add(name string, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	addToRouter(r.routerGroups[0], name, ANY, handleFunc, middlewareFunc...)
}

func (r *router) Get(name string, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	addToRouter(r.routerGroups[0], name, GET, handleFunc, middlewareFunc...)
}

func (r *router) Post(name string, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	addToRouter(r.routerGroups[0], name, POST, handleFunc, middlewareFunc...)
}

func addToRouter(root *routerGroup, name string, method MethodType, handleFunc HandleFunc, middlewareFunc ...MiddlewareFunc) {
	if root.handleFuncMap[name] == nil {
		root.handleFuncMap[name] = make(MethodHandleMap, 0)
	}
	root.handleFuncMap[name][method] = &HandlePkg{}
	pkg := root.handleFuncMap[name][method]
	pkg.handleFunc = handleFunc
	pkg.middlewareFunc = append(pkg.middlewareFunc, middlewareFunc...)
	root.treeNode.Put(name)
}

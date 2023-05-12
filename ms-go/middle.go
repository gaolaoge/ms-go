package ms_go

/*
中间件的作用是以可插拔的方式给应用添加额外的功能，但不影响原有执行。
中间件是一个调用链条，所以在处理真正的业务前可能会经过多个中间件。
1. 中间件不可耦合在用户内容中；
2. 独立存在，可拿到上下文，可做出影响；

中间件分为通用中间件和独立的中间件。
*/

type MiddlewareFunc func(handleFunc HandleFunc) HandleFunc

func (rg *routerGroup) Use(middlewareFunc ...MiddlewareFunc) {
	rg.middlewareFuncSlice = append(rg.middlewareFuncSlice, middlewareFunc...)
}

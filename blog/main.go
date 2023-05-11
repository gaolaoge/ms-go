package main

import (
	"fmt"
	msgo "github.com/gaolaoge/ms-go"
)

func main() {
	engine := msgo.New()

	engine.Add("/hello", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "hello world")
	})

	engine.Get("/hello2", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "hello world2")
	})

	engine.Post("/user/create/:cord", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, ctx.R.PostFormValue("cord"))
	})

	v1 := engine.Group("/v1")
	v1.Get("/name", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "get gaoge")
	})
	v1.Post("/name", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "post gaoge")
	})
	v1.Post("/name/:id", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, ctx.R.FormValue("id"))
	})

	engine.Run()
	fmt.Println("next")
}

func init() {

}

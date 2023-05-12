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

	v1.Use(func(next msgo.HandleFunc) msgo.HandleFunc {
		return func(ctx *msgo.Context) {
			fmt.Println("middle pre")
			next(ctx)
			fmt.Println("middle post")
		}
	}, func(next msgo.HandleFunc) msgo.HandleFunc {
		return func(ctx *msgo.Context) {
			fmt.Println("middle pre2")
			next(ctx)
			fmt.Println("middle post2")
		}
	})

	v1.Get("/name", func(ctx *msgo.Context) {
		fmt.Println("handle /v1/name")
		fmt.Fprintf(ctx.W, "get gaoge")
	}, func(next msgo.HandleFunc) msgo.HandleFunc {
		return func(ctx *msgo.Context) {
			fmt.Println("sole middle pre")
			next(ctx)
			fmt.Println("sole middle post")
		}
	}, func(next msgo.HandleFunc) msgo.HandleFunc {
		return func(ctx *msgo.Context) {
			fmt.Println("sole middle pre2")
			next(ctx)
			fmt.Println("sole middle post2")
		}
	})
	v1.Post("/name", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "post gaoge")
	})
	v1.Post("/name/:id", func(ctx *msgo.Context) {
		id := ctx.R.PostFormValue("id")
		fmt.Println("id: ", id)
		fmt.Fprintf(ctx.W, id)
	})

	engine.Run()
	fmt.Println("next")
}

func init() {

}

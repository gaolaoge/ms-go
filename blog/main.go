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

	v1 := engine.Group("/v1")
	v1.Get("/name", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "get gaoge")
	})
	v1.Post("/name", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "post gaoge")
	})

	engine.Run()
	fmt.Println("next")
}

func init() {

}

package main

import (
	"fmt"

	msgo "github.com/gaolaoge/ms-go"
	"github.com/gaolaoge/ms-go/rpc"
)

func main() {
	engine := msgo.Default()

	client := rpc.NewHttpClient()

	group := engine.Group("/order")
	group.Get("/find", func(ctx *msgo.Context) {
		// 通过商品中心查询商品的信息
		body, err := client.Get("http://localhost:9002/goods/find", map[string]any{"name": "gaoge", "age": 18})
		if err != nil {
			info := fmt.Sprintf("MSRPC_ERROR: %v", err)
			panic(info)
		}
		value := fmt.Sprintf("MSRPC_VALUE: %v", string(body))
		ctx.JSON(200, value)
	})

	engine.Run(":9003")
}

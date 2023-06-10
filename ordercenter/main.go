package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gaolaogui/ordercenter/service"

	"github.com/gaolaogui/ordercenter/model"

	msgo "github.com/gaolaoge/ms-go"
	"github.com/gaolaoge/ms-go/rpc"
)

func main() {
	engine := msgo.Default()

	client := rpc.NewHttpClient()
	client.RegisterHttpService("goods", &service.GoodsService{})

	group := engine.Group("/order")
	group.Get("/find", func(ctx *msgo.Context) {
		// 通过商品中心查询商品的信息
		body, err := client.Get("http://localhost:9002/goods/find", map[string]any{"name": "gaoge", "age": 18})
		if err != nil {
			info := fmt.Sprintf("MSRPC_ERROR: %v", err)
			panic(info)
		}
		value := fmt.Sprintf("MSRPC_VALUE: %v", string(body))
		ctx.JSON(http.StatusOK, value)
	})

	group.Get("/find2", func(ctx *msgo.Context) {
		body, err := client.Do("goods", "Find").(*service.GoodsService).Find(map[string]any{"name": "gaoge", "age": 18})
		if err != nil {
			info := fmt.Sprintf("MSRPC_ERROR: %v", err)
			panic(info)
		}
		result := &model.Result{}
		json.Unmarshal(body, result)
		ctx.JSON(http.StatusOK, result)
	})

	engine.Run(":9003")
}

package main

import (
	"net/http"

	msgo "github.com/gaolaoge/ms-go"
	"github.com/gaolaogui/goodscenter/model"
)

func main() {
	engine := msgo.Default()

	group := engine.Group("/goods")
	group.Get("/find", func(ctx *msgo.Context) {
		goods := &model.Goods{Id: "1000", Name: "商品示例"}
		ctx.JSON(http.StatusOK, &model.Result{Code: 200, Msg: "成功", Data: goods})
	})

	engine.Run(":9002")
}

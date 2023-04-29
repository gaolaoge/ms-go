package main

import (
	"fmt"
	msgo "github.com/gaolaoge/ms-go"
	"net/http"
)

func main() {
	engine := msgo.New()

	engine.Add("/hello", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(writer, "hello world")
	})

	v1 := engine.Group("/v1")
	v1.Add("/name", func(writer http.ResponseWriter, requset *http.Request) {
		fmt.Fprintf(writer, "gaoge")
	})

	engine.Run()
	fmt.Println("next")
}

func init() {

}

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

	engine.Get("/hello2", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(writer, "hello world2")
	})

	v1 := engine.Group("/v1")
	v1.Get("/name", func(writer http.ResponseWriter, requset *http.Request) {
		fmt.Fprintf(writer, "get gaoge")
	})
	v1.Post("/name", func(writer http.ResponseWriter, requset *http.Request) {
		fmt.Fprintf(writer, "post gaoge")
	})

	engine.Run()
	fmt.Println("next")
}

func init() {

}

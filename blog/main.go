package main

import (
	"fmt"
	msgo "github.com/gaolaoge/ms-go"
	"net/http"
)

func main() {
	engine := msgo.New()

	engine.Add("/", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(writer, "hello world")
	})

	engine.Run()
	fmt.Println("next")
}

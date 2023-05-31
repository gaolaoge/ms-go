package main

import (
	"fmt"
	"net/http"
	"strings"

	msgo "github.com/gaolaoge/ms-go"
	msConfig "github.com/gaolaoge/ms-go/config"
	msLog "github.com/gaolaoge/ms-go/log"
)

func main() {
	return
	engine := msgo.New()
	//logger := msLog.New()
	logger := msLog.Default()
	logger.Level = msLog.LevelDebug

	//logger.Outs = append(logger.Outs, msLog.FileWriter("./log/log.log"))
	logger.SetLogPath("./log")

	engine.Add("/hello", func(ctx *msgo.Context) {
		logger.WithFields(map[string]any{"name": "gaoge", "age": 18}).Info("hello Info")
		logger.Debug("hello Debug")
		logger.Error("hello Error")
		fmt.Fprintf(ctx.W, "hello world")
	})

	engine.Get("/hello2", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "hello world2")
	})

	engine.Post("/user/create/:cord", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, ctx.R.PostFormValue("cord"))
	})

	v1 := engine.Group("/v1")

	v1.Use(msgo.Logging)

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

	v2 := engine.Group("/file")
	v2.Get("/html", func(c *msgo.Context) {
		c.HTML(200, "<h1>gaoge2</h1>")
	})

	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	v2.Get("/htmlTemp", func(c *msgo.Context) {
		_ = c.HTMLTemplate("index.html", User{"gaoge", 18}, "tpl/index.html", "tpl/header.html")
	})

	v2.Get("/htmlTempGlob", func(c *msgo.Context) {
		_ = c.HTMLTemplateGlob("index.html", User{"gaoge", 28}, "tpl/*.html")
	})

	engine.LoadTemplate("tpl/*.html")
	v2.Get("/template", func(c *msgo.Context) {
		_ = c.Template("index.html", User{"gaoge", 1228})
	})

	v2.Get("/json", func(c *msgo.Context) {
		data := User{"gaoge", 20}
		c.JSON(200, data)
	})

	v2.Get("/xml", func(c *msgo.Context) {
		data := User{"gaoge", 20}
		c.XML(200, data)
	})

	v2.Get("/excel", func(c *msgo.Context) {
		c.File("tpl/test.xlsx")
	})

	v2.Get("/excelName", func(c *msgo.Context) {
		c.FileAttachment("tpl/test.xlsx", "demo.xlsx")
	})

	v2.Get("/jpeg", func(c *msgo.Context) {
		c.FileAttachment("tpl/xiaoliuya.jpeg", "demo.jpeg")
	})

	v2.Get("/fs", func(c *msgo.Context) {
		c.FileFromFS("xiaoliuya.jpeg", http.Dir("tpl"))
	})

	v2.Get("/redirect", func(c *msgo.Context) {
		c.Redirect(200, "/hello")
	})

	v2.Get("/string", func(c *msgo.Context) {
		c.String(http.StatusOK, "和 %s 一起 %s 。\n", "asx", "xs")
	})

	v3 := engine.Group("/params")
	v3.Use(msgo.Logging)

	v3.Get("/getQuery", func(c *msgo.Context) {
		name := c.GetQuery("name")
		c.HTML(http.StatusOK, name)
	})

	v3.Get("/getQuerys", func(c *msgo.Context) {
		name, _ := c.GetQueryArray("name")
		c.HTML(http.StatusOK, strings.Join(name, ","))
	})

	v3.Post("/file", func(ctx *msgo.Context) {
		//file := ctx.FormFile("file")
		//ctx.SaveUploadFile(file, "./upload/"+file.Filename)
		form, _ := ctx.MultipartForm()
		fmt.Println(form)
		ctx.W.Write([]byte("success"))

	})

	v3.Post("/files", func(ctx *msgo.Context) {
		files, _ := ctx.FormFiles("img")
		for _, file := range files {
			ctx.SaveUploadFile(file, "./upload/"+file.Filename)
		}
		ctx.W.Write([]byte("success"))
	})

	engine.Run()
	fmt.Println("next")
}

func init() {
	fmt.Println(msConfig.Config{})
}

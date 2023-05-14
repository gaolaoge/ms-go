package ms_go

import (
	"net/http"
	"net/url"
	"text/template"

	"github.com/gaolaoge/ms-go/render"

	utils "github.com/gaolaoge/ms-go/utils"
)

type Context struct {
	W      http.ResponseWriter
	R      *http.Request
	engine *Engine
}

func (c *Context) HTML(status int, html string) error {
	return c.Render(status, &render.HTML{Data: html, IsTemplate: false})
}

// HTMLTemplate template.ParseFiles 接收若干「文件路径」
// 模版及模版内引入的其它模版都需要被导入
func (c *Context) HTMLTemplate(name string, data any, filenames ...string) error {
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	temp := template.New(name)
	temp, err := temp.ParseFiles(filenames...)
	if err != nil {
		return err
	}
	err = temp.Execute(c.W, data)
	return err
}

// HTMLTemplateGlob template.ParseGlob 接收所需文件的「通配路径」
func (c *Context) HTMLTemplateGlob(name string, data any, pattern string) error {
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	temp := template.New(name)
	temp, err := temp.ParseGlob(pattern)
	if err != nil {
		return err
	}
	err = temp.Execute(c.W, data)
	return err
}

// Template 如果使用到模板，并不需要在访问的时候再加载，
// 可以在启动的时候，就将所有的模板加载到内存中，这样加快访问速度
func (c *Context) Template(name string, data any) error {
	return c.Render(http.StatusOK, &render.HTML{
		Data:       data,
		Name:       name,
		IsTemplate: true,
		Template:   c.engine.HTMLRender.Template,
	})
}

func (c *Context) JSON(status int, data any) error {
	return c.Render(status, &render.JSON{Data: data})
}

func (c *Context) XML(status int, data any) error {
	return c.Render(status, &render.XML{Data: data})
}

// File 文件下载
func (c *Context) File(filename string) {
	http.ServeFile(c.W, c.R, filename)
}

// FileAttachment 文件下载，自定义文件名
func (c *Context) FileAttachment(filepath, filename string) {
	if utils.IsASCII(filename) {
		c.W.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	} else {
		c.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8 "`+url.QueryEscape(filename)+`"`)
	}
	http.ServeFile(c.W, c.R, filepath)
}

// FileFromFS filepath 是相对文件系统的路径
func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		c.R.URL.Path = old
	}(c.R.URL.Path)

	c.R.URL.Path = filepath
	http.FileServer(fs).ServeHTTP(c.W, c.R)
}

func (c *Context) Redirect(code int, path string) error {
	return c.Render(code, &render.Redirect{Location: path, Request: c.R, Code: code})
}

func (c *Context) String(status int, format string, values ...any) error {
	return c.Render(status, &render.String{Format: format, Data: values})
}

func (c Context) Render(statusCode int, r render.Render) error {
	c.W.WriteHeader(statusCode)
	err := r.Render(c.W)
	return err
}

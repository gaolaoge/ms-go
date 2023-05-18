package ms_go

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"text/template"

	"github.com/gaolaoge/ms-go/render"

	utils "github.com/gaolaoge/ms-go/utils"
)

const defaultMultipartMemory = 32 << 20

type Context struct {
	W                     http.ResponseWriter
	R                     *http.Request
	engine                *Engine
	queryCache            url.Values
	formCache             url.Values
	DisallowUnknownFields bool
	IsValidate            bool
}

func (c Context) DealJson(obj any) error {
	valueOf := reflect.ValueOf(obj)
	// 判断实体值类型，若不为指针则直接报错
	if valueOf.Kind() != reflect.Pointer {
		return errors.New("this argument must have a pointer type")
	}
	// post 的参数内容在 body 中
	body := c.R.Body
	if body == nil {
		return errors.New("invalid request")
	}
	decoder := json.NewDecoder(body)
	if c.DisallowUnknownFields {
		decoder.DisallowUnknownFields() // 校验参数，若存在未知参数即报错
	}
	if c.IsValidate {
		err := validateRequireParam(obj, decoder) // 校验参数，若缺少定义参数即报错
		if err != nil {
			return err
		}
		return nil
	} else {
		return decoder.Decode(obj)
	}
}

func (c *Context) FormFile(name string) *multipart.FileHeader {
	file, header, err := c.R.FormFile(name)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	return header
}

func (c *Context) FormFiles(name string) ([]*multipart.FileHeader, error) {
	multipartForm, err := c.MultipartForm()
	if err != nil {
		return make([]*multipart.FileHeader, 0), err
	}
	return multipartForm.File[name], nil
}

func (c *Context) GetQuery(key string) string {
	c.initQueryCache()
	return c.queryCache.Get(key)
}

func (c *Context) GetQueryArray(key string) ([]string, bool) {
	c.initQueryCache()
	values, ok := c.queryCache[key]
	return values, ok
}

func (c *Context) GetDefaultQuery(key, defaultValue string) string {
	val, ok := c.GetQueryArray(key)
	if ok {
		return val[0]
	}
	return defaultValue
}

func (c *Context) initQueryCache() {
	if c.R != nil {
		c.queryCache = c.R.URL.Query()
	} else {
		c.queryCache = url.Values{}
	}
}

func (c *Context) SaveUploadFile(file *multipart.FileHeader, dst string) {
	src, err := file.Open()
	if err != nil {
		log.Println("1", err)
		return
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		log.Println("2", err)
		return
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	if err != nil {
		log.Println("3", err)
		return
	}
}

func (c *Context) GetPostForm(key string) string {
	c.initPostFormCache()
	return c.formCache.Get(key)
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.R.ParseMultipartForm(defaultMultipartMemory)
	return c.R.MultipartForm, err
}

func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	c.initPostFormCache()
	values, ok := c.formCache[key]
	return values, ok
}

func (c *Context) initPostFormCache() {
	if c.R != nil {
		if err := c.R.ParseMultipartForm(defaultMultipartMemory); err != nil {
			if !errors.Is(err, http.ErrNotMultipart) {
				log.Println(err)
				return
			}
		}
		c.formCache = c.R.PostForm
	} else {
		c.formCache = url.Values{}
	}
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

func (c *Context) Render(statusCode int, r render.Render) error {
	if statusCode != http.StatusOK {
		c.W.WriteHeader(statusCode)
	}
	err := r.Render(c.W)
	return err
}

func validateRequireParam(data any, decoder *json.Decoder) error {
	if data == nil {
		return nil
	}
	// 1. 得到指针实参对应的 *reflect.Value
	valueOf := reflect.ValueOf(data)
	// 2. 借用 *reflect.Value 得到指针指向的值再得到其接口
	elem := valueOf.Elem().Interface()
	// 3. 得到接口的 reflect.Value
	of := reflect.ValueOf(elem)
	// 4. 判断该 reflect.Value 的基础类型
	switch of.Kind() {
	case reflect.Struct:
		return checkParam(data, of, decoder)
	case reflect.Slice, reflect.Array:
	// TODO 尚未支持循环验证
	default:
		_ = decoder.Decode(data)
	}
	return nil
}

func checkParam(data any, of reflect.Value, decoder *json.Decoder) error {
	// 首先将结构体解析为map ，然后对比其 key
	// 需判断其为 结构体，才能对其转换为 map
	mapData := make(map[string]interface{})
	// TODO DisallowUnknownFields校验失效
	_ = decoder.Decode(&mapData)
	for i := 0; i < of.NumField(); i++ {
		field := of.Type().Field(i)
		name := field.Tag.Get("json")
		require := field.Tag.Get("require")
		if name == "" {
			name = field.Name
		}
		value := mapData[name]
		if value == nil && require == "true" {
			return errors.New(fmt.Sprintf("filed [%s] is not exist", name))
		}
	}
	// 5. 重新将值从 map 转为 JSON
	marshal, _ := json.Marshal(mapData)
	// 6. 将 JSON 值传入 目标
	return json.Unmarshal(marshal, data)
	//_ = decoder.Decode(data)
}

package binding

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

type jsonBinding struct {
	DisallowUnknownFields bool
	IsValidate            bool
}

func (j *jsonBinding) Name() string { return "" }

func (j *jsonBinding) Bind(r *http.Request, obj any) error {
	// post 的参数内容在 body 中
	body := r.Body
	if body == nil {
		return errors.New("invalid request")
	}
	decoder := json.NewDecoder(body)
	if j.DisallowUnknownFields {
		decoder.DisallowUnknownFields() // 校验参数，若存在未知参数即报错
	}
	if j.IsValidate {
		err := validateRequireParam(obj, decoder) // 校验参数，若缺少定义参数即报错
		if err != nil {
			return err
		}
		return nil
	} else {
		return decoder.Decode(obj)
	}

	return nil
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

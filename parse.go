package easy_api

import (
	"bytes"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io"
	"net/http"
	url2 "net/url"
	"reflect"
	"strings"
)

type IParse interface {
	Parse(interface{}, string, string) (*http.Request, error)
}

type PareRequestArgs uint8

func NewPareRequestArgs() *PareRequestArgs{
	var object PareRequestArgs = 0
	return &object
}

// Parse 解析http参数
func (p *PareRequestArgs) Parse(req interface{}, method, url string) (*http.Request, error) {
	value := reflect.ValueOf(req)
	types := reflect.TypeOf(req)
	if value.Kind() == reflect.Ptr{
		value = value.Elem()
		types = types.Elem()
	}

	queries := url2.Values{}
	var json map[string]interface{}
	for i:=0; i<value.NumField(); i++{
		arg := value.Field(i)

		if arg.Kind() == reflect.Ptr{
			arg = arg.Elem()
		}

		argType := types.Field(i)
		param := argType.Tag.Get("param")
		if param != ""{
			url = strings.Replace(url, ":"+param, fmt.Sprintf("%v", arg.Interface()), 1)
			continue
		}
		
		query := argType.Tag.Get("query")
		if query != ""{
			queries.Add(query, fmt.Sprintf("%v", arg.Interface()))
			continue
		}

		jsons := argType.Tag.Get("json")
		if jsons != ""{
			if json == nil{
				json = make(map[string]interface{})
			}
			json[jsons] = arg.Interface()
			continue
		}
	}

	var buff io.Reader
	if json != nil{
		data, err := jsoniter.Marshal(req)
		if err != nil{
			return nil, err
		}

		buff = bytes.NewBuffer(data)
	}

	httpRequest, err := http.NewRequest(method, url+"?"+queries.Encode(), buff)
	if err != nil{
		return nil, err
	}

	return httpRequest, nil
}

package easy_api

import (
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
)

type IEncode interface {
	Encode(*http.Response, interface{}) error
}

type JsonEncode uint8

func NewJsonEncode() *JsonEncode{
	var object JsonEncode = 0
	return &object
}

func (* JsonEncode) Encode(src *http.Response, dst interface{}) error  {
	buff, err := ioutil.ReadAll(src.Body)
	if err != nil{
		return err
	}

	if err = jsoniter.Unmarshal(buff, dst); err != nil{
		return err
	}

	return nil
}

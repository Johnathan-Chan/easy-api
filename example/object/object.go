package main

import (
	easy_api "github.com/Johnathan-Chan/easy-api"
	"log"
	"net/http"
)

type GetSheetInfoRequest struct {
	SpreadsheetToken string `param:"SpreadsheetToken"`
	ExtFields        string `query:"ext_fields"`
	UserIdType       string `query:"user_id_type"`
}

func SetGlobalMiddleware(req *http.Request){
	// 默认 Header Content-Type
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	//// 添加 access_token
	//req.Header.Set("Authorization", "Bearer "+ "123123123")
	//
	//// 添加 User-Agent
	//req.Header.Set("User-Agent", "fastwego/feishu")
}

func main() {
	client := easy_api.NewClient(easy_api.NewPareRequestArgs())

	// 注册注解
	client.Register(http.MethodGet, "https://www.baidu.com", GetSheetInfoRequest{})

	// 请求中间件
	client.UseRequest(SetGlobalMiddleware)

	//// 响应解析器
	jsonPares := easy_api.NewJsonEncode()

	var result string
 	if err := client.Do(GetSheetInfoRequest{
		SpreadsheetToken: "shtcnylYjqdPEfyLpNXSpJeSjGc",
	}, &result, jsonPares); err != nil{
		panic(err)
	}

	log.Println(result)
}



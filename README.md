# 利用注解快速封装api

```shell
go get github.com/Johnathan-Chan/easy-api
```

```go
// GetSheetInfoRequest @Request(method="GET", url="https://open.feishu.cn/open-apis/sheets/v2/spreadsheets/:SpreadsheetToken/metainfo")
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

	// 添加 access_token
	req.Header.Set("Authorization", "Bearer "+ "123123123")

	// 添加 User-Agent
	req.Header.Set("User-Agent", "fastwego/feishu")
}

func main() {
	client := easy_api.NewClient(easy_api.NewPareRequestArgs())

	// 注册注解
	if err := client.Register("./example"); err != nil{
		panic(err)
	}
	// 请求中间件
	client.UseRequest(SetGlobalMiddleware)

	// 响应解析器
	jsonPares := easy_api.NewJsonEncode()

	// 响应对象
	result := make(map[string]interface{})
	
	// 请求
	if err := client.Do(GetSheetInfoRequest{
		SpreadsheetToken: "shtcnylYjqdPEfyLpNXSpJeSjGc",
	}, &result, jsonPares); err != nil{
		panic(err)
	}

	log.Println(result)
}
```

 - 封装api只需要在请求的对象添加注解
```go
// @Request 注解
// method 请求方法
// url 请求路径
// GetSheetInfoRequest @Request(method="GET", url="https://open.feishu.cn/open-apis/sheets/v2/spreadsheets/:SpreadsheetToken/metainfo")
type GetSheetInfoRequest struct {
	SpreadsheetToken string `param:"SpreadsheetToken"` // param 会被解析为url中对应的参数
	ExtFields        string `query:"ext_fields"`       // query 会被解析为?之后的查询参数
	UserIdType       string `query:"user_id_type"`     // json  会解析为body
}
```
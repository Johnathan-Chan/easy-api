package easy_api

import (
	"errors"
	"net/http"
	"reflect"
)

const (
	ANNOTION_REQUEST = "@Request(.*)"
	METHOD           = "method=\"[^\"]*\""
	URL              = "url=\"[^\"]*\""
)

var Method = map[string]struct{}{
	http.MethodGet:     {},
	http.MethodHead:    {},
	http.MethodPost:    {},
	http.MethodPut:     {},
	http.MethodPatch:   {},
	http.MethodDelete:  {},
	http.MethodConnect: {},
	http.MethodOptions: {},
	http.MethodTrace:   {},
}

type RequestHandler func(*http.Request)
type ResponseHandler func(*http.Response)

type request struct {
	Method string
	Url    string
}

var (
	parseFail  = errors.New("parse failed")
	methodFail = errors.New("the method must be in GET HEAD POST PUT PATCH DELETE CONNECT OPTIONS TRACE")
	urlFail    = errors.New("url must be not null")
)

type IClient interface {
	Register(method, path string, object interface{})
	Do(interface{}, interface{}, IEncode) error
	Middleware
}

type Middleware interface {
	UseRequest(...RequestHandler)
	UseResponse(...ResponseHandler)
}

type Client struct {
	httpClient      *http.Client
	annotationCache map[string]*request
	beforeRequest   []RequestHandler
	afterResponse   []ResponseHandler
	IParse
}

func NewClient(parse IParse) IClient {
	return &Client{
		httpClient:      http.DefaultClient,
		annotationCache: make(map[string]*request),
		IParse:          parse,
	}
}

func (c *Client) UseRequest(handler ...RequestHandler) {
	c.beforeRequest = append(c.beforeRequest, handler...)
}

func (c *Client) UseResponse(handler ...ResponseHandler) {
	c.afterResponse = append(c.afterResponse, handler...)
}


func (c *Client) Register(method, path string, object interface{}) {
	objectType := reflect.TypeOf(object)
	if objectType.Kind() == reflect.Ptr {
		objectType = objectType.Elem()
	}
	c.annotationCache[objectType.Name()] = &request{Method: method, Url: path}
}



func (c *Client) Do(req, resp interface{}, encode IEncode) error {
	// 请求解析 req
	value := reflect.ValueOf(req)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	requestStruct, ok := c.annotationCache[value.Type().Name()]
	if !ok {
		return parseFail
	}

	httpRequest, err := c.Parse(req, requestStruct.Method, requestStruct.Url)
	if err != nil {
		return err
	}

	// 执行请求中间件
	for _, beforeRequest := range c.beforeRequest{
		beforeRequest(httpRequest)
	}
	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return err
	}
	if httpResponse.Body != nil {
		defer httpResponse.Body.Close()
	}

	// 执行响应中间件
	for _, afterResponse := range c.afterResponse{
		afterResponse(httpResponse)
	}
	if err = encode.Encode(httpResponse, resp); err != nil {
		return err
	}

	return nil
}

var JsonCode = NewJsonEncode()

func (c *Client) JsonDo(req, resp interface{}) error {
	return c.Do(req, resp, JsonCode)
}

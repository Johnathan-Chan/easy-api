package easy_api

import (
	"errors"
	"go/ast"
	"go/token"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
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
	Register(string) error
	Do(interface{}, interface{}, IEncode) error
	Middleware
}

type Middleware interface {
	UseRequest(...RequestHandler)
	NextRequestHandler(*http.Request)
	UseResponse(...ResponseHandler)
	NextResponseHandler(*http.Response)
}

type Client struct {
	httpClient   *http.Client
	annotationCache map[string]*request
	beforeIndex int
	afterIndex int
	beforeRequest []RequestHandler
	afterResponse []ResponseHandler
	IParse
}

func NewClient() IClient{
	return &Client{
		httpClient: http.DefaultClient,
		annotationCache: make(map[string]*request),
		IParse: NewPareRequestArgs(),
		beforeRequest: make([]RequestHandler, 1),
		afterResponse: make([]ResponseHandler, 1),
	}
}

func (c *Client) UseRequest(handler ...RequestHandler)  {
	c.beforeRequest = append(c.beforeRequest, handler...)
}

func (c *Client) NextRequestHandler(req *http.Request){
	c.beforeIndex++
	for c.beforeIndex < len(c.beforeRequest){
		c.beforeRequest[c.beforeIndex](req)
		c.beforeIndex++
	}
}

func (c *Client) UseResponse(handler ...ResponseHandler)  {
	c.afterResponse = append(c.afterResponse, handler...)
}

func (c *Client) NextResponseHandler(resp *http.Response)  {
	c.afterIndex++
	for c.afterIndex < len(c.afterResponse){
		c.afterResponse[c.afterIndex](resp)
		c.afterIndex++
	}
}

func (c *Client) Register(pkgPath string) error {
	fileSet := token.NewFileSet()
	goRegex := regexp.MustCompile("(.go)$")
	annotationRegex := regexp.MustCompile(ANNOTION_REQUEST)
	methodRegex := regexp.MustCompile(METHOD)
	urlRegex := regexp.MustCompile(URL)

	if err := filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
		if !goRegex.MatchString(info.Name()) {
			return nil
		}

		decls, err := ScanComments(path, fileSet)
		if err != nil {
			return err
		}

		for _, decl := range decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				return parseFail
			}

			// 解析注解
			for _, comment := range genDecl.Doc.List {
				if annotationRegex.MatchString(comment.Text) {
					annotation := annotationRegex.FindString(comment.Text)
					method := methodRegex.FindString(annotation)
					method = strings.Replace(method, "\"", "", -1)
					method = strings.Replace(method, "method=", "", -1)
					_, ok := Method[method]
					if !ok {
						log.Println(ok)
						return methodFail
					}

					url := urlRegex.FindString(annotation)
					url = strings.Replace(url, "\"", "", -1)
					url = strings.Replace(url, "url=", "", -1)
					if url == "" {
						return urlFail
					}

					for _, v := range genDecl.Specs {
						typeSpec, ok := v.(*ast.TypeSpec)
						if !ok {
							return parseFail
						}

						c.annotationCache[typeSpec.Name.Obj.Name] = &request{
							Method: method,
							Url:    url,
						}
					}
				}
			}

		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (c *Client) Do(req, resp interface{}, encode IEncode) error  {
	// 请求解析 req
	value := reflect.ValueOf(req)
	if value.Kind() == reflect.Ptr{
		value = value.Elem()
	}
	requestStruct, ok := c.annotationCache[value.Type().Name()]
	if !ok {
		return parseFail
	}

	httpRequest, err := c.Parse(req, requestStruct.Method, requestStruct.Url)
	if err != nil{
		return err
	}

	// 执行请求中间件
	c.NextRequestHandler(httpRequest)
	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return err
	}
	if httpResponse.Body != nil{
		defer httpResponse.Body.Close()
	}

	// 执行响应中间件
	c.NextResponseHandler(httpResponse)
	if err = encode.Encode(httpResponse, resp); err != nil{
		return err
	}

	return nil
}
package eggo

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/binding"
)

type H map[string]interface{}

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	// 请求信息
	Path   string // 路径
	Method string // 方法
	Params map[string]string
	// 响应信息
	StatusCode int
	// 中间件
	handlers []HandlerFunc
	index    int
	// Engine 指针
	engine *Engine
	//读写上下文的互斥锁
	mutex sync.RWMutex
	//用户写入的信息
	Keys map[string]interface{}
}

//构造一个新的中间件
func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Path:    req.URL.Path,
		Method:  req.Method,
		Request: req,
		Writer:  w,
		index:   -1,
	}
}

func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

//抛弃请求
func (c *Context) Abort() {
	c.index = len(c.handlers)
}

//携带状态抛弃请求
func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.Abort()
}

//绑定JSON
func (c *Context) Bind(obj interface{}) error {
	return binding.JSON.Bind(c.Request, obj)
}

//写上下文
func (c *Context) Set(key string, value interface{}) {
	c.mutex.Lock()
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
	c.mutex.Unlock()
}

//读取上下文
func (c *Context) Get(key string) (value interface{}, exists bool) {
	c.mutex.Lock()
	value, exists = c.Keys[key]
	c.mutex.Unlock()
	return
}

func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if c.Request.MultipartForm == nil {
		if err := c.Request.ParseMultipartForm(c.engine.MaxMultipartMemory); err != nil {
			return nil, err
		}
	}
	f, fh, err := c.Request.FormFile(name)
	if err != nil {
		return nil, err
	}
	f.Close()
	return fh, err
}

//保存上传的文件
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

//获取客户端ip
func (c *Context) ClientIP() string {
	ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err != nil {
		return ""
	}
	remoteIP := net.ParseIP(ip)
	return remoteIP.String()
}

func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) PostForm(key string) string {
	return c.Request.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) DefaultQuery(key, defaultValue string) string {
	if value := c.Query(key); len(value) != 0 {
		return value
	}
	return defaultValue
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

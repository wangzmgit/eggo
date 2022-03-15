package eggo

import (
	"net/http"
	"path"
)

type RouterGroup struct {
	prefix   string
	handlers []HandlerFunc // 路由组的中间件
	parent   *RouterGroup  // 所属的组
	engine   *Engine
}

//创建一个新的RouterGroup
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix:   group.prefix + prefix,
		handlers: group.handlers,
		parent:   group,
		engine:   engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// 将中间件添加到组中
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.handlers = append(group.handlers, middlewares...)
}

//添加路由
func (group *RouterGroup) addRoute(method string, comp string, handlers []HandlerFunc) {
	pattern := group.prefix + comp
	finalSize := len(group.handlers) + len(handlers)
	mergedHandlers := make([]HandlerFunc, finalSize)
	copy(mergedHandlers, group.handlers)
	copy(mergedHandlers[len(group.handlers):], handlers)
	group.engine.router.addRoute(method, pattern, mergedHandlers)
	printRoute(method, pattern, mergedHandlers) //输出路由信息
}

// 添加 OPTIONS 请求
func (group *RouterGroup) OPTIONS(pattern string, handlers ...HandlerFunc) {
	group.addRoute("OPTIONS", pattern, handlers)
}

// 添加 HEAD 请求
func (group *RouterGroup) HEAD(pattern string, handlers ...HandlerFunc) {
	group.addRoute("HEAD", pattern, handlers)
}

// 添加 GET 请求
func (group *RouterGroup) GET(pattern string, handlers ...HandlerFunc) {
	group.addRoute("GET", pattern, handlers)
}

// 添加 POST 请求
func (group *RouterGroup) POST(pattern string, handlers ...HandlerFunc) {
	group.addRoute("POST", pattern, handlers)
}

// 添加 PUT 请求
func (group *RouterGroup) PUT(pattern string, handlers ...HandlerFunc) {
	group.addRoute("PUT", pattern, handlers)
}

// 添加 DELETE 请求
func (group *RouterGroup) DELETE(pattern string, handlers ...HandlerFunc) {
	group.addRoute("DELETE", pattern, handlers)
}

// 创建静态处理程序
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// 检查文件是否存在或是否有权访问
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

// 静态文件的请求
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// 注册 GET 处理程序
	group.GET(urlPattern, handler)
}

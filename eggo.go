package eggo

import (
	"net/http"
)

// 处理函数，参数为上下文
type HandlerFunc func(*Context)

// Engine实现ServeHTTP的接口
type Engine struct {
	*RouterGroup
	router             *router
	groups             []*RouterGroup // 路由组
	MaxMultipartMemory int64          //上传文件的最大内存
}

// Engine的构造函数
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{
		engine: engine,
	}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// 默认返回已附加 Logger 和 Recovery 中间件的 Engine 实例。
func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

// 启动服务并侦听和处理 HTTP 请求。
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := newContext(w, req)
	ctx.engine = engine
	engine.router.handle(ctx)
}

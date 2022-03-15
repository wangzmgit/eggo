package eggo

import (
	"fmt"
	"reflect"
	"runtime"
	"time"
)

//输出路由信息
func printRoute(method string, pattern string, handlers []HandlerFunc) {
	handlersLen := len(handlers)
	handlerName := nameOfFunction(getLastHandler(handlers, handlersLen))
	fmt.Printf("[eggo] %-6s %-25s --> %s (%d handlers)\n", method, pattern, handlerName, handlersLen)
}

//输出请求信息
func printRequest(method string, elapsed time.Duration, StatusCode int, ip, pattern string) {
	time := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[eggo] %-20s | %-7s | %-9s | %-3d | %-10s | %s\n", time, method, elapsed, StatusCode, ip, pattern)
}

//获取函数名
func nameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

//获取最后的处理函数
func getLastHandler(handlers []HandlerFunc, length int) HandlerFunc {
	return handlers[length-1]
}

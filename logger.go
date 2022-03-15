package eggo

import (
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		start := time.Now()
		c.Next()
		// 打印请求信息
		printRequest(c.Method, time.Since(start), c.StatusCode, c.ClientIP(), c.Path)
	}
}

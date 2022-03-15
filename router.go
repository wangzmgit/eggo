package eggo

import (
	"net/http"
	"strings"
)

type router struct {
	roots    *node
	handlers map[string][]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    &node{},
		handlers: make(map[string][]HandlerFunc),
	}
}

func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")

	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

//添加路由
func (r *router) addRoute(method, pattern string, handlers []HandlerFunc) {
	parts := parsePattern(pattern)
	r.roots.insert(method, pattern, parts, 0)
	r.handlers[pattern] = handlers
}

func (r *router) getRoute(path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root := r.roots

	n := root.search(searchParts, 0)

	if n != nil {
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}

	return nil, nil
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Path)
	if n != nil {
		c.Params = params
		c.handlers = append(c.handlers, r.handlers[n.pattern]...)
	} else {
		//找不到请求的处理
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND\n")
		})
	}
	c.Next()
}

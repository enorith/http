package router

import (
	"bytes"
	"strings"
	"sync"

	"github.com/enorith/http/contracts"
	"github.com/enorith/http/pipeline"
)

const (
	GET     = 1
	HEAD    = 1 << 1
	POST    = 1 << 2
	PUT     = 1 << 3
	PATCH   = 1 << 4
	DELETE  = 1 << 5
	OPTIONS = 1 << 6
	ANY     = GET | HEAD | POST | PUT | PATCH | DELETE | OPTIONS
)

var methodMap = map[int]string{
	GET:     "GET",
	HEAD:    "HEAD",
	POST:    "POST",
	PUT:     "PUT",
	PATCH:   "PATCH",
	DELETE:  "DELETE",
	OPTIONS: "OPTIONS",
}

//RouteHandler normal route handler
type RouteHandler func(r contracts.RequestContract) contracts.ResponseContract

type partial struct {
	segment []byte
	isParam bool
}

type ParamRoute struct {
	path       string
	partials   []partial
	handler    RouteHandler
	middleware []string
	isValid    bool
	pipeFuncs  []pipeline.PipeFunc
	name       string
}

func (p *ParamRoute) SetMiddleware(middleware []string) *ParamRoute {
	p.middleware = middleware
	return p
}

func (p *ParamRoute) Middleware() []string {
	return p.middleware
}

func (p *ParamRoute) IsValid() bool {
	return p.isValid
}

//Partials partials of route path
func (p *ParamRoute) Partials() []partial {
	return p.partials
}

//Handler route handler
func (p *ParamRoute) Handler() RouteHandler {
	return p.handler
}

//Path path
func (p *ParamRoute) Path() string {
	return p.path
}

//PipeFuncs return pipeline middleware
func (p *ParamRoute) PipeFuncs() []pipeline.PipeFunc {
	return p.pipeFuncs
}

//Name of route
func (p *ParamRoute) Name() string {
	return p.name
}

type routesHolder struct {
	routes []*ParamRoute
}

func (rh *routesHolder) Middleware(middleware ...string) *routesHolder {
	for _, v := range rh.routes {
		ms := append(v.middleware, middleware...)
		v.SetMiddleware(ms)
	}
	return rh
}

func (rh *routesHolder) Use(p ...pipeline.PipeFunc) *routesHolder {
	for _, v := range rh.routes {
		ps := append(v.pipeFuncs, p...)
		v.pipeFuncs = ps
	}
	return rh
}

func (rh *routesHolder) Name(name string) *routesHolder {
	for _, v := range rh.routes {
		v.name = name
	}
	return rh
}

type methodTrees struct {
	mu    *sync.RWMutex
	nodes map[string]*node
}

func (mt *methodTrees) get(method string) *node {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	return mt.nodes[method]
}

func (mt *methodTrees) set(method string, n *node) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	mt.nodes[method] = n
}

type router struct {
	routes map[string][]*ParamRoute
	trees  *methodTrees
	prefix string
}

func (r *router) Routes() map[string][]*ParamRoute {
	return r.routes
}

//HandleGet get method with route handler
func (r *router) HandleGet(path string, handler RouteHandler) *routesHolder {
	return r.Register(GET, path, handler)
}

func (r *router) HandlePost(path string, handler RouteHandler) *routesHolder {
	return r.Register(POST, path, handler)
}

func (r *router) HandlePut(path string, handler RouteHandler) *routesHolder {
	return r.Register(PUT, path, handler)
}

func (r *router) HandlePatch(path string, handler RouteHandler) *routesHolder {
	return r.Register(PATCH, path, handler)
}

func (r *router) HandleDelete(path string, handler RouteHandler) *routesHolder {
	return r.Register(DELETE, path, handler)
}

//Register register route
func (r *router) Register(method int, path string, handler RouteHandler) *routesHolder {
	var routes []*ParamRoute
	for i := GET; i <= OPTIONS; i <<= 1 {
		if m, ok := methodMap[i]; i&method > 0 && ok {
			routes = append(routes, r.addRoute(m, path, handler))
		}
	}

	return &routesHolder{
		routes,
	}
}

func (r *router) addRoute(method string, path string, handler RouteHandler) *ParamRoute {
	path = JoinPaths(r.prefix, path)
	route := &ParamRoute{
		path:    path,
		handler: handler,
		isValid: true,
	}

	r.routes[method] = append(r.routes[method], route)
	tree := r.trees.get(method)
	if tree == nil {
		tree = new(node)
		r.trees.set(method, tree)
	}
	tree.addRoute(path, route)

	return route
}

func (r *router) Match(request contracts.RequestContract) *ParamRoute {
	// using bytes
	return r.MatchTree(request)
}

func (r *router) MatchTree(request contracts.RequestContract) *ParamRoute {
	method := request.GetMethod()
	sp := string(r.normalPath(request.GetPathBytes()))
	tree := r.trees.get(method)

	if tree != nil {
		value := tree.getValue(sp)
		if value.route != nil {
			request.SetParams(value.params)
			request.SetParamsSlice(value.paramsSlice)
			request.SetRouteName(value.route.name)
			return value.route
		}
	}

	return &ParamRoute{}
}

// Depracated MatchBytes, using MatchTree
//
func (r *router) MatchBytes(request contracts.RequestContract) *ParamRoute {
	method := request.GetMethod()
	pathBytes := r.normalPath(request.GetPathBytes())
	bytesPartials := bytes.Split(pathBytes, []byte("/"))

	partialLength := len(bytesPartials)

	for _, route := range r.routes[method] {
		/// static match
		if bytes.Equal([]byte(route.path), pathBytes) {
			return route
		} else if len(route.partials) == partialLength {
			/// /test/foo -> /test/:name

			/// path bytes partials
			/// {test, foo}

			params := map[string][]byte{}
			var paramsSlice [][]byte
			matches := 0

			for index, part := range bytesPartials {

				pa := route.partials[index].segment

				if route.partials[index].isParam {
					/// is parameter route
					/// pa = :name part=foo
					params[string(pa[1:])] = part
					paramsSlice = append(paramsSlice, part)
					matches++
				} else if bytes.Equal(pa, part) {
					/// pa = test part=test
					matches++
				}
			}
			if matches == partialLength {
				request.SetParams(params)
				request.SetParamsSlice(paramsSlice)
				return route
			}
		}
	}

	return &ParamRoute{
		isValid: false,
	}
}

// Depracated MatchString, using MatchTree
//
func (r *router) MatchString(request contracts.RequestContract) *ParamRoute {
	method := request.GetMethod()
	sp := string(r.normalPath(request.GetPathBytes()))
	partials := strings.Split(sp, "/")
	l := len(partials)

	for _, route := range r.routes[method] {
		if route.path == sp {
			return route
		} else if len(route.partials) == l {
			params := map[string][]byte{}
			var paramsSlice [][]byte

			matches := 0
			/// /test/foo => /test/:name
			for index, part := range partials {

				/// is parameter
				pa := route.partials[index].segment

				if route.partials[index].isParam {
					params[string(pa[1:])] = []byte(part)
					paramsSlice = append(paramsSlice, []byte(part))
					matches++
				} else if bytes.Equal(pa, []byte(part)) {
					matches++
				}
			}
			if matches == l {
				request.SetParams(params)
				request.SetParamsSlice(paramsSlice)
				return route
			}
		}
	}

	return &ParamRoute{
		isValid: false,
	}
}

// trim last "/"
func (r *router) normalPath(path []byte) []byte {
	l := len(path)

	if l > 1 && path[l-1] == '/' {
		return path[0 : l-1]
	}

	return path
}

func AppendSlash(path string) string {
	if strings.Index(path, "/") != 0 {
		path = "/" + path
	}

	return path
}

func JoinPaths(paths ...string) string {
	var ps []string
	for _, path := range paths {
		prefix := strings.TrimPrefix(path, "/")
		if prefix != "" {
			ps = append(ps, prefix)
		}
	}

	return AppendSlash(strings.Join(ps, "/"))
}

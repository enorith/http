package router

import (
	"bytes"
	"github.com/enorith/http/contracts"
	"strings"
	"sync"
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

type paramRoute struct {
	path       string
	partials   []partial
	handler    RouteHandler
	middleware []string
	isParam    bool
	isValid    bool
}

func (p *paramRoute) SetMiddleware(middleware []string) {
	p.middleware = middleware
}

func (p *paramRoute) Middleware() []string {
	return p.middleware
}

func (p *paramRoute) IsValid() bool {
	return p.isValid
}

//Partials partials of route path
func (p *paramRoute) Partials() []partial {
	return p.partials
}

//Handler route handler
func (p *paramRoute) Handler() RouteHandler {
	return p.handler
}

//Path path
func (p *paramRoute) Path() string {
	return p.path
}

type routesHolder struct {
	routes []*paramRoute
}

func (rh *routesHolder) Middleware(middleware ...string) *routesHolder {
	for _, v := range rh.routes {
		v.SetMiddleware(middleware)
	}
	return rh
}

func (rh *routesHolder) Prefix(prefix string) *routesHolder {
	if strings.Index(prefix, "/") != 0 {
		prefix = "/" + prefix
	}

	for _, v := range rh.routes {
		v.path = prefix + v.path
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
	routes map[string][]*paramRoute
	trees  *methodTrees
}

func (r *router) Routes() map[string][]*paramRoute {
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
	var routes []*paramRoute
	for i := GET; i <= OPTIONS; i <<= 1 {
		if m, ok := methodMap[i]; i&method > 0 && ok {
			routes = append(routes, r.addRoute(m, path, handler))
		}
	}

	return &routesHolder{
		routes,
	}
}

func (r *router) addRoute(method string, path string, handler RouteHandler) *paramRoute {
	route := &paramRoute{
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

func resolvePartials(path string) []partial {
	ps := bytes.Split([]byte(path), []byte("/"))
	var partials []partial
	for _, v := range ps {
		if bytes.HasPrefix(v, []byte(":")) {
			partials = append(partials, partial{
				segment: v,
				isParam: true,
			})
		} else {
			partials = append(partials, partial{
				segment: v,
				isParam: false,
			})
		}
	}
	return partials
}

func (r *router) Match(request contracts.RequestContract) *paramRoute {
	// using bytes
	return r.MatchTree(request)
}

func (r *router) MatchTree(request contracts.RequestContract) *paramRoute {
	method := request.GetMethod()
	sp := string(r.normalPath(request.GetPathBytes()))
	tree := r.trees.get(method)

	if tree != nil {
		value := tree.getValue(sp)
		if value.route != nil {
			request.SetParams(value.params)
			request.SetParamsSlice(value.paramsSlice)
			return value.route
		}
	}

	return &paramRoute{}
}

func (r *router) MatchBytes(request contracts.RequestContract) *paramRoute {
	method := request.GetMethod()
	pathBytes := r.normalPath(request.GetPathBytes())
	bytesPartials := bytes.Split(pathBytes, []byte("/"))

	partialLength := len(bytesPartials)

	for _, route := range r.routes[method] {
		/// static match
		if bytes.Compare([]byte(route.path), pathBytes) == 0 {
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
				} else if bytes.Compare(pa, part) == 0 {
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

	return &paramRoute{
		isValid: false,
	}
}

///
func (r *router) MatchString(request contracts.RequestContract) *paramRoute {
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
				} else if bytes.Compare(pa, []byte(part)) == 0 {
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

	return &paramRoute{
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

func (r *router) hashRequestRoute(rq contracts.RequestContract) string {
	return string(rq.GetUri())
}

package router

import (
	"bytes"
	"github.com/enorith/http/contracts"
	"strings"
)

const (
	GET    = 1
	HEAD   = 1 << 1
	POST   = 1 << 2
	PUT    = 1 << 3
	PATCH  = 1 << 4
	DELETE = 1 << 5
	OPTION = 1 << 6
	ANY    = GET | HEAD | POST | PUT | PATCH | DELETE | OPTION
)

var methodMap = map[int]string{
	GET:    "GET",
	HEAD:   "HEAD",
	POST:   "POST",
	PUT:    "PUT",
	PATCH:  "PATCH",
	DELETE: "DELETE",
	OPTION: "OPTION",
}

//RouteHandler normal route handler
type RouteHandler func(r contracts.RequestContract) contracts.ResponseContract

type pathPartial [2]string

type paramRoute struct {
	path       string
	partials   []pathPartial
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
func (p *paramRoute) Partials() []pathPartial {
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
	for _, v := range rh.routes {
		v.path = prefix + v.path
	}
	return rh
}

type router struct {
	routes map[string][]*paramRoute
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
	for i := GET; i <= DELETE; i <<= 1 {
		if m, ok := methodMap[i]; i&method > 0 && ok {
			var route *paramRoute
			if strings.Contains(path, "/:") {
				route = r.addParamRoute(m, path, handler)
			} else {
				route = r.addRoute(m, path, handler)
			}
			routes = append(routes, route)
		}
	}
	return &routesHolder{
		routes,
	}
}

func (r *router) addRoute(method string, path string, handler RouteHandler) *paramRoute {
	router := &paramRoute{
		path:    path,
		handler: handler,
		isParam: false,
		isValid: true,
	}
	r.routes[method] = append(r.routes[method], router)

	return router
}

func (r *router) addParamRoute(method string, path string, handler RouteHandler) *paramRoute {
	partials := resolvePartials(path)
	router := &paramRoute{
		partials: partials,
		path:     path,
		handler:  handler,
		isParam:  true,
		isValid:  true,
	}
	r.routes[method] = append(r.routes[method], router)

	return router
}

func resolvePartials(path string) []pathPartial {
	ps := strings.Split(path, "/")
	var partials []pathPartial
	for _, v := range ps {
		if strings.HasPrefix(v, ":") {
			partials = append(partials, [2]string{v, "1"})
		} else {
			partials = append(partials, [2]string{v, "0"})
		}
	}
	return partials
}

func (r *router) Match(request contracts.RequestContract) *paramRoute {
	// using bytes
	return r.MatchBytes(request)
}

func (r *router) MatchBytes(request contracts.RequestContract) *paramRoute {
	method := request.GetMethod()
	pathBytes := r.normalPath(request.GetPathBytes())

	for _, v := range r.routes[method] {
		/// full match
		if bytes.Compare([]byte(v.path), pathBytes) == 0 {
			return v
		}
	}

	/// /test/foo -> /test/:name

	/// path bytes partials
	/// {test, foo}
	bytesPartials := bytes.Split(pathBytes, []byte("/"))

	partialLength := len(bytesPartials)

	for _, route := range r.routes[method] {
		/// same amount of partials
		/// route.partials -> {test, :name}
		if len(route.partials) == partialLength {
			params := map[string][]byte{}
			var paramsSlice [][]byte
			matches := 0

			for index, part := range bytesPartials {

				pa := route.partials[index][0]

				if route.partials[index][1] == "1" {
					/// is parameter route
					/// pa = :name part=foo
					params[pa[1:]] = part
					paramsSlice = append(paramsSlice, part)
					matches++
				} else if bytes.Compare([]byte(pa), part) == 0 {
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

	for _, v := range r.routes[method] {
		if v.path == sp {
			return v
		}
	}

	partials := strings.Split(sp, "/")
	l := len(partials)
	/// Match parameter route
	for _, route := range r.routes[method] {
		/// same amount of partials
		if len(route.partials) == l {
			params := map[string][]byte{}
			matches := 0
			/// /test/foo => /test/:name
			for index, part := range partials {

				/// is parameter
				pa := route.partials[index][0]

				if route.partials[index][1] == "1" {
					params[pa[1:]] = []byte(part)
					matches++
				} else if pa == part {
					matches++
				}
			}
			if matches == l {
				request.SetParams(params)
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

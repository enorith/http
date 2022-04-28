package pipeline

import (
	"fmt"

	"github.com/enorith/http/contracts"
)

//RequestMiddleware request middleware
type RequestMiddleware interface {
	Handle(r contracts.RequestContract, next PipeHandler) contracts.ResponseContract
}

type MiddlewareGroup map[string][]RequestMiddleware

type middlewareChain []RequestMiddleware

func (mc middlewareChain) Handle(r contracts.RequestContract, next PipeHandler) contracts.ResponseContract {
	pipe := new(Pipeline)

	pipe.Send(r)
	for _, m := range mc {
		pipe.ThroughMiddleware(m)
	}
	return pipe.Then(next)
}

func MiddlewareChain(mid ...RequestMiddleware) middlewareChain {
	return mid
}

type ResponsePipe struct {
	middleware []RequestMiddleware
}

func (rp *ResponsePipe) Pipe(pipeFunc PipeFunc) *ResponsePipe {

	rp.middleware = append(rp.middleware, FuncMiddleware{HandleFunc: pipeFunc})
	fmt.Printf("pipe %p\n", rp)

	return rp
}

func (rp *ResponsePipe) Handle(r contracts.RequestContract, next PipeHandler) contracts.ResponseContract {
	fmt.Printf("handle %p\n", rp)

	return MiddlewareChain(rp.middleware...).Handle(r, next)
}

type FuncMiddleware struct {
	HandleFunc PipeFunc
}

func (fm FuncMiddleware) Handle(r contracts.RequestContract, next PipeHandler) contracts.ResponseContract {
	return fm.HandleFunc(r, next)
}

func NewResponsePipe() *ResponsePipe {
	return &ResponsePipe{
		middleware: make([]RequestMiddleware, 0),
	}
}

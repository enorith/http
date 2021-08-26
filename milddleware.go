package http

import "github.com/enorith/http/contracts"

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

package http

import "github.com/enorith/http/contracts"

//PipeHandler destination handler
type PipeHandler func(r contracts.RequestContract) contracts.ResponseContract

//PipeFunc request middleware function
type PipeFunc func(r contracts.RequestContract, next PipeHandler) contracts.ResponseContract

//Pipeline is request pipeline prepare for request middleware
type Pipeline struct {
	pipes []PipeFunc
	r     contracts.RequestContract
}

//Send request to pipeline
func (p *Pipeline) Send(r contracts.RequestContract) *Pipeline {
	p.r = r
	return p
}

//Through middleware
func (p *Pipeline) Through(pipe interface{}) *Pipeline {
	if p.pipes == nil {
		p.pipes = []PipeFunc{}
	}
	p.pipes = append(p.pipes, p.preparePipe(pipe))

	return p
}

//ThroughFunc through middleware function
func (p *Pipeline) ThroughFunc(pipe PipeFunc) *Pipeline {
	p.Through(pipe)
	return p
}

//ThroughMiddleware through middleware struct
func (p *Pipeline) ThroughMiddleware(pipe RequestMiddleware) *Pipeline {
	p.Through(pipe)
	return p
}

//Then final destination
func (p *Pipeline) Then(handler PipeHandler) contracts.ResponseContract {

	return func(r contracts.RequestContract) contracts.ResponseContract {
		if p.pipes != nil && len(p.pipes) > 0 {
			next := p.prepareNext(0, handler)
			return p.pipes[0](r, next)
		}
		return handler(r)
	}(p.r)
}

func (p *Pipeline) prepareNext(now int, handler PipeHandler) PipeHandler {
	l := len(p.pipes)
	var next PipeHandler
	if now+1 >= l {
		next = handler
	} else {
		next = func(r contracts.RequestContract) contracts.ResponseContract {
			return p.pipes[now+1](r, p.prepareNext(now+1, handler))
		}
	}

	return next
}

func (p *Pipeline) preparePipe(pipe interface{}) PipeFunc {

	if t, ok := pipe.(PipeFunc); ok {
		return t
	}

	if t, ok := pipe.(RequestMiddleware); ok {
		return func(r contracts.RequestContract, next PipeHandler) contracts.ResponseContract {
			return t.Handle(r, next)
		}
	}

	return func(r contracts.RequestContract, next PipeHandler) contracts.ResponseContract {
		return next(r)
	}
}

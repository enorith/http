package http

import (
	"bytes"
	"fmt"
	"github.com/enorith/container"
	"github.com/enorith/exception"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/errors"
	"github.com/enorith/http/router"
	"github.com/enorith/http/validation"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"reflect"
	"time"
)

const Version = "v0.0.1"

type handlerType int

const DefaultConcurrency = 256 * 1024

const (
	HandlerFastHttp handlerType = iota
	HandlerNetHttp
)

//RequestMiddleware request middleware
type RequestMiddleware interface {
	Handle(r contracts.RequestContract, next PipeHandler) contracts.ResponseContract
}

type MiddlewareGroup map[string][]RequestMiddleware

func timeMic() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}

type Kernel struct {
	wrapper            *router.Wrapper
	middleware         []RequestMiddleware
	middlewareGroup    map[string][]RequestMiddleware
	errorHandler       errors.ErrorHandler
	tcpKeepAlive       bool
	RequestCurrency    int
	MaxRequestBodySize int
	OutputLog          bool
	Handler            handlerType
}

func (k *Kernel) Wrapper() *router.Wrapper {
	return k.wrapper
}

func (k *Kernel) handleFunc(f func() (request contracts.RequestContract, code int)) {
	var start int64
	if k.OutputLog {
		start = timeMic()
	}
	request, code := f()

	if k.OutputLog {
		end := timeMic()

		body := bytes.Join(bytes.Fields(request.GetContent()), []byte(""))
		log.Printf("/ %s - [%s] %s '%s' (%d) <%.3fms>", request.GetClientIp(),
			request.GetMethod(), request.GetUri(), body, code, float64(end-start)/1000)
	}
}

func (k *Kernel) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	k.handleFunc(func() (request contracts.RequestContract, code int) {
		request = content.NewNetHttpRequest(r, w)
		resp := k.Handle(request)

		if resp != nil {
			if k.tcpKeepAlive {
				resp.SetHeader("Connection", "keep-alive")
			}

			headers := resp.Headers()
			if headers != nil {
				for k, v := range headers {
					w.Header().Set(k, v)
				}
			}
			if !resp.Handled() {
				// call after set headers, before write body
				w.WriteHeader(resp.StatusCode())
			}
			body := resp.Content()
			if tp, ok := resp.(*content.TemplateResponse); ok {
				temp := tp.Template()
				temp.Execute(w, tp.TemplateData())
			} else if tp, ok := resp.(*content.File); ok {
				http.ServeFile(w, r, tp.Path())
			} else if body != nil {
				w.Write(body)
			}
			code = resp.StatusCode()
		}

		return
	})
}

func (k *Kernel) FastHttpHandler(ctx *fasthttp.RequestCtx) {
	k.handleFunc(func() (request contracts.RequestContract, code int) {
		request = content.NewFastHttpRequest(ctx)
		resp := k.Handle(request)

		if k.tcpKeepAlive {
			resp.SetHeader("Connection", "keep-alive")
		}

		ctx.Response.SetStatusCode(resp.StatusCode())
		if resp.Headers() != nil {
			for k, v := range resp.Headers() {
				ctx.Response.Header.Set(k, v)
			}
		}
		if tp, ok := resp.(*content.TemplateResponse); ok {
			temp := tp.Template()
			temp.Execute(ctx, tp.TemplateData())
		} else if tp, ok := resp.(*content.File); ok {
			fasthttp.ServeFile(ctx, tp.Path())
		} else {
			body := resp.Content()
			buf := bytes.NewBuffer(body)

			fmt.Fprint(ctx, buf)
		}
		code = resp.StatusCode()

		return
	})
}

func (k *Kernel) SetMiddlewareGroup(middlewareGroup map[string][]RequestMiddleware) {
	k.middlewareGroup = middlewareGroup
}

func (k *Kernel) SetMiddleware(ms []RequestMiddleware) {
	k.middleware = ms
}

func (k *Kernel) KeepAlive(b ...bool) *Kernel {
	if len(b) > 0 {
		k.tcpKeepAlive = b[0]
	} else {
		k.tcpKeepAlive = true
	}
	return k
}

func (k *Kernel) IsKeepAlive() bool {
	return k.tcpKeepAlive
}

func (k *Kernel) SetErrorHandler(handler errors.ErrorHandler) {
	k.errorHandler = handler
}

func (k *Kernel) Handle(r contracts.RequestContract) (resp contracts.ResponseContract) {
	// End-able request
	rc := make(chan contracts.ResponseContract)

	go func() {
		defer func() {
			if x := recover(); x != nil {
				resp = k.errorHandler.HandleError(x, r)
				r.End(resp)
			}
		}()
		r.End(k.SendRequestToRouter(r))
	}()

	go func() {
		select {
		case <- r.Ended():
			rc <- r.GetResponse()
			return
		}
	}()

	resp = <-rc

	if t, ok := resp.(*content.ErrorResponse); ok {
		resp = k.errorHandler.HandleError(t.E(), r)
	}

	if t, ok := resp.(exception.Exception); ok {
		resp = k.errorHandler.HandleError(t, r)
	}

	return resp
}

func (k *Kernel) SendRequestToRouter(r contracts.RequestContract) contracts.ResponseContract {
	pipe := new(Pipeline)
	pipe.Send(r)
	for _, m := range k.middleware {
		pipe.ThroughMiddleware(m)
	}
	p := k.wrapper.Match(r)
	if !p.IsValid() {
		return content.NotFoundResponse("not found")
	}
	if mid := p.Middleware(); mid != nil {
		for _, v := range mid {
			if ms, exists := k.middlewareGroup[v]; exists {
				for _, md := range ms {
					pipe.ThroughMiddleware(md)
				}
			}
		}
	}

	return pipe.Then(func(r contracts.RequestContract) contracts.ResponseContract {
		//resp := k.wrapper.Dispatch(r)
		return p.Handler()(r)
	})
}

func NewKernel(cr router.ContainerRegister, debug bool) *Kernel {
	k := new(Kernel)
	k.wrapper = router.NewWrapper(cr)
	k.wrapper.ResolveRequest(KernelRequestResolver{})
	k.errorHandler = &errors.StandardErrorHandler{
		Debug: debug,
	}
	k.RequestCurrency = DefaultConcurrency
	return k
}

type KernelRequestResolver struct {

}

func (rr KernelRequestResolver) ResolveRequest(r contracts.RequestContract, runtime *container.Container) {
	runtime.RegisterSingleton(r)
	runtime.Singleton("contracts.RequestContract", r)

	runtime.BindFunc(&content.Request{}, func(c *container.Container) reflect.Value {

		return reflect.ValueOf(&content.Request{RequestContract: r})
	}, false)

	runtime.BindFunc(content.Request{}, func(c *container.Container) reflect.Value {

		return reflect.ValueOf(content.Request{RequestContract: r})
	}, false)

	runtime.WithInjector(RequestInjector{runtime: runtime, request: r, validator: validation.DefaultValidator})
}

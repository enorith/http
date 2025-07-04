package http

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/enorith/container"
	"github.com/enorith/exception"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/errors"
	"github.com/enorith/http/pipeline"
	"github.com/enorith/http/router"
	"github.com/valyala/fasthttp"
)

const Version = "v1.2.4"

type handlerType int

const DefaultConcurrency = 256 * 1024

const (
	HandlerFastHttp handlerType = iota
	HandlerNetHttp
)

var RequestLogger = func(request contracts.RequestContract, statusCode int, start time.Time) {
	log.Printf("/ %s - [%s] %s (%d) <%s>", request.RemoteAddr(),
		request.GetMethod(), request.GetUri(), statusCode, time.Since(start))
}

type RequestResolver interface {
	ResolveRequest(r contracts.RequestContract, container container.Interface)
}

type ContainerRegister func(request contracts.RequestContract) container.Interface

type Kernel struct {
	wrapper            *router.Wrapper
	middleware         []pipeline.RequestMiddleware
	middlewareGroup    map[string][]pipeline.RequestMiddleware
	errorHandler       errors.ErrorHandler
	tcpKeepAlive       bool
	RequestCurrency    int
	MaxRequestBodySize int
	OutputLog          bool
	Handler            handlerType
	cr                 ContainerRegister
	resolver           RequestResolver
}

func (k *Kernel) Wrapper() *router.Wrapper {
	return k.wrapper
}

func (k *Kernel) handleFunc(f func() (request contracts.RequestContract, code int)) {
	start := time.Now()
	request, code := f()

	if k.OutputLog && RequestLogger != nil {
		RequestLogger(request, code, start)
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

			resp.SetHeader("Server", fmt.Sprintf("enorith/%s (net/http)", Version))

			headers := resp.Headers()
			for k, v := range headers {
				w.Header().Set(k, v)
			}
			if cr, ok := resp.(contracts.WithResponseCookies); ok {
				for _, c := range cr.Cookies() {
					http.SetCookie(w, c)
				}
			}
			if !resp.Handled() {
				// call after set headers, before write body
				w.WriteHeader(resp.StatusCode())
			}

			if tp, ok := resp.(contracts.TemplateResponseContract); ok {
				temp := tp.Template()
				temp.Execute(w, tp.TemplateData())
			} else if fp, ok := resp.(*content.File); ok {
				http.ServeFile(w, r, fp.Path())
			} else if wp, ok := resp.(io.WriterTo); ok {
				wp.WriteTo(w)
			} else {
				body := resp.Content()
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
		if rr, ok := resp.(*content.RedirectResponse); ok {
			ctx.Redirect(rr.URL(), rr.StatusCode())
		}

		ctx.Response.SetStatusCode(resp.StatusCode())
		if resp.Headers() != nil {
			for k, v := range resp.Headers() {
				ctx.Response.Header.Set(k, v)
			}
		}
		if cr, ok := resp.(contracts.WithResponseCookies); ok {
			for _, c := range cr.Cookies() {
				ctx.Response.Header.Add("Set-Cookie", c.String())
			}
		}
		ctx.Response.Header.Set("Server", fmt.Sprintf("enorith/%s (fasthttp)", Version))
		if fs, ok := resp.(*content.FastHttpFileServer); ok {
			h := GetFsHandler(fs.Root(), fs.StripSlashes())
			h(ctx)
			return
		}

		if tp, ok := resp.(contracts.TemplateResponseContract); ok {
			temp := tp.Template()
			temp.Execute(ctx, tp.TemplateData())
		} else if sr, ok := resp.(*content.StreamResponse); ok {
			strem := sr.Stream()
			defer strem.Close()
			io.Copy(ctx, strem)
		} else if fp, ok := resp.(*content.File); ok {
			fasthttp.ServeFile(ctx, fp.Path())
		} else if wp, ok := resp.(io.WriterTo); ok {
			wp.WriteTo(ctx)
		} else {
			body := resp.Content()
			buf := bytes.NewBuffer(body)

			fmt.Fprint(ctx, buf)
		}
		code = resp.StatusCode()

		return
	})
}

func (k *Kernel) SetMiddlewareGroup(middlewareGroup map[string][]pipeline.RequestMiddleware) {
	k.middlewareGroup = middlewareGroup
}

func (k *Kernel) SetMiddleware(ms []pipeline.RequestMiddleware) {
	k.middleware = ms
}

func (k *Kernel) Use(m pipeline.RequestMiddleware) *Kernel {
	k.middleware = append(k.middleware, m)
	return k
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
	defer func() {
		if x := recover(); x != nil {
			resp = k.errorHandler.HandleError(x, r, true)
		}
	}()
	ioc := k.cr(r)
	r.SetContainer(ioc)
	k.resolver.ResolveRequest(r, ioc)

	resp = k.SendRequestToRouter(r)

	if t, ok := resp.(*content.ErrorResponse); ok {
		resp = k.errorHandler.HandleError(t, r, false)
	}

	if t, ok := resp.(exception.Exception); ok {
		resp = k.errorHandler.HandleError(t, r, false)
	}

	return resp
}

func (k *Kernel) SendRequestToRouter(r contracts.RequestContract) contracts.ResponseContract {
	pipe := new(pipeline.Pipeline)
	pipe.Send(r)
	for _, m := range k.middleware {
		pipe.ThroughMiddleware(m)
	}
	p := k.wrapper.Match(r)
	if !p.IsValid() {
		return content.NotFoundResponse("not found")
	}
	pfs := p.PipeFuncs()
	for _, pf := range pfs {
		pipe.Through(pf)
	}

	ioc := r.GetContainer()
	ioc.Bind(&router.ParamRoute{}, p, true)

	mid := p.Middleware()
	for _, v := range mid {
		if ms, exists := k.middlewareGroup[v]; exists {
			for _, md := range ms {
				pipe.ThroughMiddleware(md)
			}
		}
		midKey := "middleware." + v
		if ioc.Bound(midKey) {
			instance, e := ioc.Instance(midKey)
			if e != nil {
				return content.ErrResponseFromError(e, 500, nil)
			}
			pipe.Through(instance.Interface())
		}
	}

	return pipe.Then(func(r contracts.RequestContract) contracts.ResponseContract {
		//resp := k.wrapper.Dispatch(r)
		return p.Handler()(r)
	})
}

func NewKernel(cr ContainerRegister, debug bool) *Kernel {
	k := new(Kernel)
	k.cr = cr
	k.resolver = KernelRequestResolver{}
	k.wrapper = router.NewWrapper()
	k.errorHandler = &errors.StandardErrorHandler{
		Debug: debug,
	}
	k.RequestCurrency = DefaultConcurrency
	k.middleware = []pipeline.RequestMiddleware{}
	k.middlewareGroup = make(map[string][]pipeline.RequestMiddleware)
	return k
}

package http

import (
	"github.com/enorith/http/router"
	"github.com/valyala/fasthttp"
	"log"
	net "net/http"
)

type RouterRegister func(ro *router.Wrapper, k *Kernel)

type Server struct {
	k *Kernel
}

func (s *Server) Serve(addr string, register RouterRegister) error {
	register(s.k.Wrapper(), s.k)
	var err error
	if s.k.Handler == HandlerFastHttp {
		log.Printf("enorith/%s (fasthttp) served at [%s]", Version, addr)
		err = s.GetFastHttpServer(s.k).
			ListenAndServe(addr)
	} else if s.k.Handler == HandlerNetHttp {
		log.Printf("enorith/%s (net/http) served at [%s]", Version, addr)
		err = net.ListenAndServe(addr, s.k)
	}

	return err
}

func (s *Server) GetFastHttpServer(kernel *Kernel) *fasthttp.Server {

	return &fasthttp.Server{
		Handler:            kernel.FastHttpHandler,
		Concurrency:        kernel.RequestCurrency,
		TCPKeepalive:       kernel.IsKeepAlive(),
		MaxRequestBodySize: kernel.MaxRequestBodySize,
	}
}

func NewServer(cr router.ContainerRegister, debug bool) *Server {
	k := NewKernel(cr, debug)

	return &Server{k: k}
}


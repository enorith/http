package http

import (
	"context"
	"fmt"
	"log"
	net "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enorith/http/router"
	"github.com/valyala/fasthttp"
)

var (
	ReadTimeout  = time.Second * 30
	WriteTimeout = time.Second * 30
	IdleTimeout  = time.Second * 30
)

type RouterRegister func(rw *router.Wrapper, k *Kernel)

type Server struct {
	k *Kernel
}

func (s *Server) Serve(addr string, register RouterRegister) {
	register(s.k.Wrapper(), s.k)

	if s.k.Handler == HandlerFastHttp {
		s.serveFastHttp(addr)
	} else if s.k.Handler == HandlerNetHttp {
		s.serveNetHttp(addr)
	}
}

func (s *Server) serveFastHttp(addr string) {
	srv := s.GetFastHttpServer(s.k)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := srv.ListenAndServe(addr); err != nil {
			log.Fatalf("listen %s error: %v", addr, err)
		}
	}()
	log.Printf("%s served at [%s]", logPrefix("fasthttp"), addr)
	<-done
	log.Printf("%s stoping...", logPrefix("fasthttp"))
	if e := srv.Shutdown(); e != nil {
		log.Fatalf("%s shutdown error: %v", logPrefix("fasthttp"), e)
	}
	log.Printf("%s stopped", logPrefix("fasthttp"))
}

func (s *Server) serveNetHttp(addr string) {
	srv := net.Server{
		Addr:         addr,
		Handler:      s.k,
		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != net.ErrServerClosed {
			log.Fatalf("listen %s error: %v", addr, err)
		}
	}()
	log.Printf("%s served at [%s]", logPrefix("net/http"), addr)
	<-done
	log.Printf("%s stoping...", logPrefix("net/http"))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if e := srv.Shutdown(ctx); e != nil {
		log.Fatalf("%s shutdown error: %v", logPrefix("net/http"), e)
	}
	log.Printf("%s stopped", logPrefix("net/http"))
}

func (s *Server) GetFastHttpServer(kernel *Kernel) *fasthttp.Server {

	return &fasthttp.Server{
		Handler:            kernel.FastHttpHandler,
		Concurrency:        kernel.RequestCurrency,
		TCPKeepalive:       kernel.IsKeepAlive(),
		MaxRequestBodySize: kernel.MaxRequestBodySize,
		ReadTimeout:        ReadTimeout,
		WriteTimeout:       WriteTimeout,
		IdleTimeout:        IdleTimeout,
	}
}

func NewServer(cr ContainerRegister, debug bool) *Server {
	k := NewKernel(cr, debug)

	return &Server{k: k}
}

func logPrefix(handler string) string {
	return fmt.Sprintf("enorith/%s (%s)", Version, handler)
}

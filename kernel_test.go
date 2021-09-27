package http_test

import (
	"testing"

	"github.com/enorith/container"
	"github.com/enorith/http"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/pipeline"
	"github.com/enorith/http/tests"
)

var k *http.Kernel

func BenchmarkKernel_Handle(b *testing.B) {
	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		k.Handle(tests.NewRequest("GET", "/hello"))
	}
}

func TestKernel_Handle(t *testing.T) {
	resp := k.Handle(tests.NewRequest("GET", "/hello"))

	t.Log("handle result", resp.StatusCode(), resp.Content())
}

func TestKernel_HandleCustom(t *testing.T) {
	resp := k.Handle(tests.NewRequest("GET", "/test"))

	t.Log("handle result", resp.StatusCode(), string(resp.Content()), resp.Headers())
}

func TestKernel_HandleInject(t *testing.T) {
	resp := k.Handle(tests.NewRequest("GET", "/inject?bar=injection data"))

	if string(resp.Content()) != "injection data" {
		t.Log("injection failed")
		t.Fail()
	}
}
func TestKernel_HandleValidate(t *testing.T) {
	resp := k.Handle(tests.NewRequest("GET", "/inject"))

	t.Log("handle result validate", resp.StatusCode(), string(resp.Content()), resp.Headers())
}

func TestKernel_Middleware(t *testing.T) {
	resp := k.Handle(tests.NewRequest("GET", "/mid"))

	t.Log("handle mid result", resp.StatusCode(), string(resp.Content()), resp.Headers())
}

func TestKernel_Middleware2(t *testing.T) {
	resp := k.Handle(tests.NewRequest("GET", "/mid2"))

	t.Log("handle mid2 result", resp.StatusCode(), string(resp.Content()), resp.Headers())
}

type CustomResp string

func (c CustomResp) StatusCode() int {
	return 422
}

func (c CustomResp) Headers() map[string]string {
	return map[string]string{
		"X-Client": string(c),
	}
}

type Foo struct {
	content.Request
	Bar string `input:"bar" validate:"required"`
}

type DemoMiddleware struct {
}

func (d DemoMiddleware) Handle(r contracts.RequestContract, next pipeline.PipeHandler) contracts.ResponseContract {
	resp := next(r)

	resp.SetHeader("X-Middleware", "demo")

	return resp
}

type Demo2Middleware struct {
}

func (d Demo2Middleware) Handle(r contracts.RequestContract, next pipeline.PipeHandler) contracts.ResponseContract {
	resp := next(r)

	resp.SetHeader("X-Middleware2", "demo2")

	return resp
}

func init() {
	k = http.NewKernel(func(request contracts.RequestContract) container.Interface {

		con := container.New()
		con.BindFunc("middleware.test", func(c container.Interface) (interface{}, error) {
			var dm DemoMiddleware

			return dm, nil
		}, false)

		con.BindFunc("middleware.test2", func(c container.Interface) (interface{}, error) {
			var dm DemoMiddleware
			var dm2 Demo2Middleware

			return pipeline.MiddlewareChain(dm, dm2), nil
		}, false)
		return con
	}, false)

	w := k.Wrapper()
	w.Get("/hello", func() []byte {
		return []byte("ok")
	})

	w.Get("/test", func() CustomResp {
		return CustomResp("test")
	})

	w.Get("/inject", func(foo Foo) string {
		return foo.Bar
	})
	w.Get("/mid", func() string {
		return "ok"
	}).Middleware("test")

	w.Get("/mid2", func() string {
		return "ok"
	}).Middleware("test2")
}

package router_test

import (
	"fmt"
	"testing"

	"github.com/enorith/container"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	router2 "github.com/enorith/http/router"
	"github.com/enorith/http/tests"
)

var router *router2.Wrapper

func BenchmarkRouter_Match(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		router.MatchBytes(NewRequest("GET", fmt.Sprintf("/foo/%d", i)))
	}
}

func BenchmarkRouter_MatchString(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		router.MatchString(NewRequest("GET", fmt.Sprintf("/foo/%d", i)))
	}
}

func BenchmarkRouter_MatchInjection(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		router.Match(&tests.FakeRequest{
			SimpleParamRequest: content.SimpleParamRequest{},
			Path:               "/injection/foo",
			Method:             "GET",
		})
	}
}

func BenchmarkRouter_MatchHelloworld(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		router.Match(NewRequest("GET", "/"))
	}
}

func BenchmarkRouter_MatchParamInjection(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		router.Match(NewRequest("GET", "/users/42"))
	}
}

func TestJoinPaths(t *testing.T) {
	t.Log(router2.JoinPaths("", "/", "bar", "/foo"))
}

func NewRequest(method, path string) *tests.FakeRequest {
	return &tests.FakeRequest{
		SimpleParamRequest: content.SimpleParamRequest{},
		Path:               path,
		Method:             method,
	}
}

func init() {
	router = router2.NewWrapper(func(contracts.RequestContract) container.Interface {

		return container.New()
	})
	router.HandleGet("/", func(r contracts.RequestContract) contracts.ResponseContract {
		return content.NewResponse([]byte("ok"), nil, 200)
	})

	router.Get("/injection/foo", func() contracts.ResponseContract {
		return content.JsonResponse(map[string]string{
			"s": "o",
		}, 200, nil)
	})

	router.Get("/users/:id", func(id content.ParamInt) contracts.ResponseContract {

		return content.TextResponse(fmt.Sprintf("ok %d", id), 200)
	})

	router.HandleGet("/foo/:id", func(r contracts.RequestContract) contracts.ResponseContract {
		r.Param("id")
		return content.TextResponse("ok", 200)
	})

}

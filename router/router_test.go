package router_test

import (
	"fmt"
	"github.com/enorith/container"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	router2 "github.com/enorith/http/router"
	"github.com/enorith/http/tests"
	"testing"
)

var router *router2.Wrapper


func BenchmarkRouter_Match(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		router.MatchBytes(&tests.FakeRequest{
			SimpleParamRequest: content.SimpleParamRequest{},
			Path:               fmt.Sprintf("/%d", i),
			Method:             "GET",
		})
	}
}
func BenchmarkRouter_MatchString(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		router.MatchString(&tests.FakeRequest{
			SimpleParamRequest: content.SimpleParamRequest{},
			Path:               fmt.Sprintf("/%d", i),
			Method:             "GET",
		})
	}
}
func BenchmarkRouter_MatchInjection(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		router.MatchString(&tests.FakeRequest{
			SimpleParamRequest: content.SimpleParamRequest{},
			Path:               "/injection",
			Method:             "GET",
		})
	}
}

func BenchmarkRouter_MatchHelloworld(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		router.MatchString(&tests.FakeRequest{
			SimpleParamRequest: content.SimpleParamRequest{},
			Path:               "/",
			Method:             "GET",
		})
	}
}


func init() {
	router = router2.NewWrapper(func() *container.Container {

		return container.New()
	})
	router.HandleGet("/", func(r contracts.RequestContract) contracts.ResponseContract {
		return content.NewResponse([]byte("ok"), nil,200)
	})
	router.Get("/injection", func() contracts.ResponseContract {
		return content.JsonResponse(map[string]string{
			"s" : "o",
		}, 200,nil)
	})

	router.HandleGet("/:id", func(r contracts.RequestContract) contracts.ResponseContract {
		r.Param("id")
		return content.TextResponse("ok", 200)
	})

}

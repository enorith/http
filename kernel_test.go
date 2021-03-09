package http_test

import (
	"github.com/enorith/container"
	"github.com/enorith/http"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/tests"
	"testing"
)

var k *http.Kernel

func BenchmarkKernel_Handle(b *testing.B) {
	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		k.Handle(&tests.FakeRequest{
			SimpleParamRequest: content.SimpleParamRequest{},
			Path:               "/hello",
			Method:             "GET",
		})
	}
}

func TestKernel_Handle(t *testing.T) {
	resp := k.Handle(&tests.FakeRequest{
		SimpleParamRequest: content.SimpleParamRequest{},
		Path:               "/hello",
		Method:             "GET",
	})

	t.Log("handle result", resp.StatusCode())
}

func init() {
	k = http.NewKernel(func(request contracts.RequestContract) *container.Container {

		return container.New()
	}, false)

	k.Wrapper().HandleGet("/hello", func(r contracts.RequestContract) contracts.ResponseContract {
		return content.TextResponse("ok", 200)
	})
}
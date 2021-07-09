package http_test

import (
	"testing"

	"github.com/enorith/container"
	"github.com/enorith/http"
	"github.com/enorith/http/contracts"
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

func init() {
	k = http.NewKernel(func(request contracts.RequestContract) container.Interface {

		return container.New()
	}, false)

	k.Wrapper().Get("/hello", func() []byte {
		return []byte("ok")
	})
}

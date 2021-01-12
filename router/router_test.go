package router_test

import (
	"context"
	"fmt"
	"github.com/enorith/container"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	router2 "github.com/enorith/http/router"
	"testing"
)

var router *router2.Wrapper

type fakeRequest struct {
	content.SimpleParamRequest
	path   string
	method string
}

func (f fakeRequest) GetValue(key ...string) contracts.InputValue {
	panic("implement me")
}

func (f fakeRequest) Context() context.Context {
	panic("implement me")
}

func (f fakeRequest) Accepts() []byte {
	panic("implement me")
}

func (f fakeRequest) ExceptsJson() bool {
	panic("implement me")
}

func (f fakeRequest) RequestWithJson() bool {
	panic("implement me")
}

func (f fakeRequest) IsXmlHttpRequest() bool {
	panic("implement me")
}

func (f fakeRequest) GetMethod() string {
	return f.method
}

func (f fakeRequest) GetPathBytes() []byte {
	return []byte(f.path)
}

func (f fakeRequest) GetUri() []byte {
	panic("implement me")
}

func (f fakeRequest) Get(key string) []byte {
	panic("implement me")
}

func (f fakeRequest) File(key string) (contracts.UploadFile, error) {
	panic("implement me")
}

func (f fakeRequest) GetInt64(key string) (int64, error) {
	panic("implement me")
}

func (f fakeRequest) GetUint64(key string) (uint64, error) {
	panic("implement me")
}

func (f fakeRequest) GetString(key string) string {
	panic("implement me")
}

func (f fakeRequest) GetInt(key string) (int, error) {
	panic("implement me")
}

func (f fakeRequest) GetClientIp() string {
	panic("implement me")
}

func (f fakeRequest) GetContent() []byte {
	panic("implement me")
}

func (f fakeRequest) Unmarshal(to interface{}) error {
	panic("implement me")
}

func (f fakeRequest) GetSignature() []byte {
	panic("implement me")
}

func (f fakeRequest) Header(key string) []byte {
	panic("implement me")
}

func (f fakeRequest) HeaderString(key string) string {
	panic("implement me")
}

func (f fakeRequest) SetHeader(key string, value []byte) contracts.RequestContract {
	panic("implement me")
}

func (f fakeRequest) SetHeaderString(key, value string) contracts.RequestContract {
	panic("implement me")
}

func (f fakeRequest) Authorization() []byte {
	panic("implement me")
}

func (f fakeRequest) BearerToken() ([]byte, error) {
	panic("implement me")
}

func BenchmarkRouter_Match(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		router.MatchBytes(&fakeRequest{
			SimpleParamRequest: content.SimpleParamRequest{},
			path:               fmt.Sprintf("/%d", i),
			method:             "GET",
		})
	}
}
func BenchmarkRouter_MatchString(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		router.MatchString(&fakeRequest{
			SimpleParamRequest: content.SimpleParamRequest{},
			path:               fmt.Sprintf("/%d", i),
			method:             "GET",
		})
	}
}

func init() {
	router = router2.NewWrapper(func() *container.Container {
		return container.New()
	})
	router.HandleGet("/:id", func(r contracts.RequestContract) contracts.ResponseContract {
		r.Param("id")
		return content.TextResponse("ok", 200)
	})
}

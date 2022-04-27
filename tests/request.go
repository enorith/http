package tests

import (
	"context"
	"net/url"

	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
)

type FakeRequest struct {
	content.SimpleParamRequest
	Path   string
	Method string
	Url    *url.URL
}

func (f FakeRequest) GetValue(key ...string) contracts.InputValue {
	return f.Get(key[0])
}

func (f FakeRequest) Context() context.Context {
	return context.Background()
}

func (f FakeRequest) Accepts() []byte {
	panic("implement me")
}

func (f FakeRequest) ExceptsJson() bool {
	return true
}

func (f FakeRequest) RequestWithJson() bool {
	return false
}

func (f FakeRequest) IsXmlHttpRequest() bool {
	panic("implement me")
}

func (f FakeRequest) GetMethod() string {
	return f.Method
}

func (f FakeRequest) GetPathBytes() []byte {
	return []byte(f.Path)
}

func (f FakeRequest) GetUri() []byte {
	return []byte(f.Path)
}

func (f FakeRequest) Get(key string) []byte {
	x0 := f.Url.Query().Get(key)
	return []byte(x0)
}

func (f FakeRequest) File(key string) (contracts.UploadFile, error) {
	panic("implement File")
}

func (f FakeRequest) GetInt64(key string) (int64, error) {
	panic("implement GetInt64")
}

func (f FakeRequest) GetUint64(key string) (uint64, error) {
	panic("implement m GetUint64")
}

func (f FakeRequest) GetString(key string) string {
	panic("implement me GetString")
}

func (f FakeRequest) GetInt(key string) (int, error) {
	panic("implement me GetInt")
}

func (f FakeRequest) GetClientIp() string {
	panic("implement me GetClientIp")
}

func (f FakeRequest) RemoteAddr() string {
	panic("implement me RemoteAddr")
}

func (f FakeRequest) GetContent() []byte {
	panic("implement me GetContent")
}

func (f FakeRequest) Unmarshal(to interface{}) error {
	panic("implement me Unmarshal")
}

func (f FakeRequest) GetSignature() []byte {
	panic("implement me GetSignature")
}

func (f FakeRequest) Header(key string) []byte {
	panic("implement me Header")
}

func (f FakeRequest) HeaderString(key string) string {
	panic("implement me HeaderString")
}

func (f FakeRequest) SetHeader(key string, value []byte) contracts.RequestContract {
	panic("implement me SetHeader")
}

func (f FakeRequest) SetHeaderString(key, value string) contracts.RequestContract {
	panic("implement me SetHeaderString")
}

func (f FakeRequest) Authorization() []byte {
	panic("implement me Authorization")
}

func (f FakeRequest) BearerToken() ([]byte, error) {
	panic("implement me BearerToken")
}

func (f FakeRequest) ToString() string {
	return "fake request"
}
func (f FakeRequest) GetURL() *url.URL {
	panic("not implemented") // TODO: Implement
}

func NewRequest(method, path string) *FakeRequest {
	url, _ := url.Parse(path)

	return &FakeRequest{
		SimpleParamRequest: content.SimpleParamRequest{},
		Path:               url.Path,
		Method:             method,
		Url:                url,
	}
}

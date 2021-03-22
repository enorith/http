package tests

import (
	"context"

	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
)

type FakeRequest struct {
	content.SimpleParamRequest
	Path   string
	Method string
}

func (f FakeRequest) GetValue(key ...string) contracts.InputValue {
	panic("implement me")
}

func (f FakeRequest) Context() context.Context {
	return context.Background()
}

func (f FakeRequest) Accepts() []byte {
	panic("implement me")
}

func (f FakeRequest) ExceptsJson() bool {
	panic("implement me")
}

func (f FakeRequest) RequestWithJson() bool {
	panic("implement me")
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
	panic("implement me")
}

func (f FakeRequest) File(key string) (contracts.UploadFile, error) {
	panic("implement me")
}

func (f FakeRequest) GetInt64(key string) (int64, error) {
	panic("implement me")
}

func (f FakeRequest) GetUint64(key string) (uint64, error) {
	panic("implement me")
}

func (f FakeRequest) GetString(key string) string {
	panic("implement me")
}

func (f FakeRequest) GetInt(key string) (int, error) {
	panic("implement me")
}

func (f FakeRequest) GetClientIp() string {
	panic("implement me")
}

func (f FakeRequest) GetContent() []byte {
	panic("implement me")
}

func (f FakeRequest) Unmarshal(to interface{}) error {
	panic("implement me")
}

func (f FakeRequest) GetSignature() []byte {
	panic("implement me")
}

func (f FakeRequest) Header(key string) []byte {
	panic("implement me")
}

func (f FakeRequest) HeaderString(key string) string {
	panic("implement me")
}

func (f FakeRequest) SetHeader(key string, value []byte) contracts.RequestContract {
	panic("implement me")
}

func (f FakeRequest) SetHeaderString(key, value string) contracts.RequestContract {
	panic("implement me")
}

func (f FakeRequest) Authorization() []byte {
	panic("implement me")
}

func (f FakeRequest) BearerToken() ([]byte, error) {
	panic("implement me")
}

func (f FakeRequest) String() string {
	return "fake request"
}

func NewRequest(method, path string) *FakeRequest {
	return &FakeRequest{
		SimpleParamRequest: content.SimpleParamRequest{},
		Path:               path,
		Method:             method,
	}
}

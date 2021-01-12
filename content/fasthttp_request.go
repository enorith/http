package content

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	. "github.com/enorith/http/contracts"
	"github.com/enorith/supports/byt"
	b "github.com/enorith/supports/byt"
	"github.com/valyala/fasthttp"
	"strconv"
)

type FastHttpRequest struct {
	SimpleParamRequest
	origin    *fasthttp.RequestCtx
	signature []byte
}

func (r *FastHttpRequest) GetMethod() string {
	return string(r.origin.Method())
}

func (r *FastHttpRequest) Context() context.Context {
	return r.origin
}

func (r *FastHttpRequest) GetPathBytes() []byte {
	return r.origin.Path()
}

func (r *FastHttpRequest) GetUri() []byte {
	return r.origin.RequestURI()
}

func (r *FastHttpRequest) IsXmlHttpRequest() bool {

	return bytes.Equal(r.origin.Request.Header.Peek("X-Requested-With"), []byte("XMLHttpRequest"))
}

func (r *FastHttpRequest) ExceptsJson() bool {
	return b.Contains(r.Accepts(), []byte("/json"), []byte("+json"))
}

func (r *FastHttpRequest) RequestWithJson() bool {
	return byt.Contains(r.Header("Content-Type"), []byte("application/json"))
}

func (r *FastHttpRequest) Accepts() []byte {
	return r.origin.Request.Header.Peek("Accept")
}

func (r *FastHttpRequest) GetClientIp() string {
	return r.origin.RemoteIP().String()
}

func (r *FastHttpRequest) File(key string) (UploadFile, error) {
	h, err := r.origin.FormFile(key)
	if err != nil {
		return nil, err
	}

	return &uploadFile{header: h}, nil
}
func (r *FastHttpRequest) Get(key string) []byte {
	query := r.origin.QueryArgs()
	if query.Has(key) {
		return query.Peek(key)
	}

	post := r.origin.PostArgs()
	if post.Has(key) {
		return post.Peek(key)
	}

	return GetJsonValue(r, key)
}

func (r *FastHttpRequest) GetInt(key string) (int, error) {
	str := string(r.Get(key))

	return strconv.Atoi(str)
}

func (r *FastHttpRequest) GetInt64(key string) (int64, error) {
	str := r.GetString(key)

	return strconv.ParseInt(str, 10, 64)
}

func (r *FastHttpRequest) GetUint64(key string) (uint64, error) {
	str := r.GetString(key)

	return strconv.ParseUint(str, 10, 64)
}

func (r *FastHttpRequest) GetString(key string) string {

	return string(r.Get(key))
}

func (r *FastHttpRequest) GetValue(key... string) InputValue {
	if len(key) > 0 {
		return r.Get(key[0])
	}

	return r.GetContent()
}

func (r *FastHttpRequest) Origin() *fasthttp.RequestCtx {
	return r.origin
}

func (r *FastHttpRequest) GetContent() []byte {
	return r.origin.Request.Body()
}

func (r *FastHttpRequest) Unmarshal(to interface{}) error {
	return json.Unmarshal(r.GetContent(), to)
}

func (r *FastHttpRequest) GetSignature() []byte {
	if len(r.signature) > 0 {
		return r.signature
	}

	h := sha1.New()
	var data []byte
	data = append(data, r.GetPathBytes()...)
	data = append(data, r.Origin().Method()...)
	data = append(data, r.Origin().Request.Header.UserAgent()...)
	data = append(data, r.Origin().Request.Header.Peek("Authorization")...)
	data = append(data, r.Origin().RemoteIP()...)
	data = append(data, r.Origin().QueryArgs().QueryString()...)
	data = append(data, r.Origin().PostArgs().QueryString()...)
	data = append(data, r.GetContent()...)

	h.Write(data)

	r.signature = h.Sum(nil)

	return r.signature
}

func (r *FastHttpRequest) Header(key string) []byte {
	return r.Origin().Request.Header.Peek(key)
}

func (r *FastHttpRequest) HeaderString(key string) string {
	return string(r.Header(key))
}

func (r *FastHttpRequest) SetHeader(key string, value []byte) RequestContract {
	r.Origin().Request.Header.SetBytesV(key, value)

	return r
}

func (r *FastHttpRequest) SetHeaderString(key, value string) RequestContract {
	r.Origin().Request.Header.Set(key, value)

	return r
}

func (r *FastHttpRequest) Authorization() []byte {
	return r.Header("Authorization")
}

func (r *FastHttpRequest) BearerToken() ([]byte, error) {
	auth := r.Authorization()

	if len(auth) < 7 {
		return nil, errors.New("invalid bearer token")
	}

	return bytes.TrimSpace(auth[6:]), nil
}

func NewFastHttpRequest(origin *fasthttp.RequestCtx) *FastHttpRequest {
	r := new(FastHttpRequest)
	r.origin = origin
	r.signature = []byte{}
	r.done = make(chan struct{})
	return r
}

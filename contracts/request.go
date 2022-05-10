package contracts

import (
	"context"
	"mime/multipart"
	"net/url"

	"github.com/buger/jsonparser"
	"github.com/enorith/container"
	"github.com/enorith/supports/byt"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type JsonHandler func(value InputValue, dataType jsonparser.ValueType)

type InputValue []byte

func (i InputValue) GetInt() (int64, error) {
	return byt.ToInt64(i)
}

func (i InputValue) GetUInt() (uint64, error) {
	return byt.ToUint64(i)
}

func (i InputValue) GetString() string {
	return string(i)
}

func (i InputValue) GetBool() (bool, error) {
	return byt.ToBool(i)
}

func (i InputValue) GetFloat() (float64, error) {
	return byt.ToFloat64(i)
}

func (i InputValue) GetValue(key string) (InputValue, jsonparser.ValueType, error) {
	v, dataType, _, err := jsonparser.Get(i, key)

	return v, dataType, err
}

func (i InputValue) Each(h JsonHandler) error {
	_, e := jsonparser.ArrayEach(i, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		h(value, dataType)
	})

	return e
}

func (i InputValue) Unmarshal(v interface{}) error {
	return json.Unmarshal(i, v)
}

//InputSource is general input source
type InputSource interface {
	Get(key string) []byte
	File(key string) (UploadFile, error)
	GetValue(key ...string) InputValue
}

type WithContainer interface {
	SetContainer(ioc container.Interface)
	GetContainer() container.Interface
}

type WithRouteName interface {
	SetRouteName(name string)
	GetRouteName() string
}

type WithPathInfo interface {
	GetURL() *url.URL
	GetPathBytes() []byte
	GetUri() []byte
}

type WithRequestCookies interface {
	CookieByte(key string) []byte
}

//RequestContract is interface of http request
type RequestContract interface {
	InputSource
	WithContainer
	WithRouteName
	WithPathInfo
	Context() context.Context
	Params() map[string][]byte
	Param(key string) string
	ParamBytes(key string) []byte
	ParamInt64(key string) (int64, error)
	ParamUint64(key string) (uint64, error)
	ParamInt(key string) (int, error)
	SetParams(params map[string][]byte)
	SetParamsSlice(paramsSlice [][]byte)
	ParamsSlice() [][]byte
	Accepts() []byte
	ExceptsJson() bool
	RequestWithJson() bool
	IsXmlHttpRequest() bool
	GetMethod() string
	GetInt64(key string) (int64, error)
	GetUint64(key string) (uint64, error)
	GetString(key string) string
	GetInt(key string) (int, error)
	GetClientIp() string
	RemoteAddr() string
	GetContent() []byte
	Unmarshal(to interface{}) error
	GetSignature() []byte
	Header(key string) []byte
	HeaderString(key string) string
	SetHeader(key string, value []byte) RequestContract
	SetHeaderString(key, value string) RequestContract
	Authorization() []byte
	BearerToken() ([]byte, error)
	ToString() string
}

type UploadFile interface {
	Save(dist string) error
	Open() (multipart.File, error)
	Close() error
	Filename() string
}

type InputScanner interface {
	ScanInput(data []byte) error
}

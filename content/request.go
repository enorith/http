package content

import (
	"fmt"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/enorith/container"
	. "github.com/enorith/http/contracts"
	"github.com/enorith/supports/byt"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Param string

func (p Param) Value() string {
	return string(p)
}

type ParamInt64 int64

func (p ParamInt64) Value() int64 {
	return int64(p)
}

type ParamInt int

func (p ParamInt) Value() int {
	return int(p)
}

type ParamUint64 uint64

func (p ParamUint64) Value() uint64 {
	return uint64(p)
}

type SimpleParamRequest struct {
	params      map[string][]byte
	paramsSlice [][]byte
	container   container.Interface
	routeName   string
}

func (shr *SimpleParamRequest) Params() map[string][]byte {
	return shr.params
}

func (shr *SimpleParamRequest) ParamsSlice() [][]byte {
	return shr.paramsSlice
}

func (shr *SimpleParamRequest) Param(key string) string {
	return string(shr.params[key])
}

func (shr *SimpleParamRequest) ParamBytes(key string) []byte {
	return shr.params[key]
}

func (shr *SimpleParamRequest) ParamInt64(key string) (int64, error) {
	param, ok := shr.params[key]
	if !ok {
		return 0, fmt.Errorf("can not get param [%s]", key)
	}

	return byt.ToInt64(param)
}

func (shr *SimpleParamRequest) ParamUint64(key string) (uint64, error) {
	param, ok := shr.params[key]
	if !ok {
		return 0, fmt.Errorf("can not get param [%s]", key)
	}

	return byt.ToUint64(param)
}

func (shr *SimpleParamRequest) ParamInt(key string) (int, error) {
	param, ok := shr.params[key]
	if !ok {
		return 0, fmt.Errorf("can not get param [%s]", key)
	}

	return strconv.Atoi(string(param))
}

func (shr *SimpleParamRequest) SetParams(params map[string][]byte) {
	shr.params = params
}

func (shr *SimpleParamRequest) SetParamsSlice(paramsSlice [][]byte) {
	shr.paramsSlice = paramsSlice
}

func (shr *SimpleParamRequest) SetContainer(ioc container.Interface) {
	shr.container = ioc
}

func (shr *SimpleParamRequest) GetContainer() container.Interface {
	return shr.container
}

func (shr *SimpleParamRequest) SetRouteName(name string) {
	shr.routeName = name
}

func (shr *SimpleParamRequest) GetRouteName() string {
	return shr.routeName
}

func GetJsonValue(r RequestContract, key string) []byte {
	if r.RequestWithJson() {
		val, _, _, _ := jsonparser.Get(r.GetContent(), key)

		return val
	}

	return nil
}

type Request struct {
	RequestContract
}

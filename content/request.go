package content

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/enorith/container"
	"github.com/enorith/http/contracts"
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

func GetJsonValue(r contracts.RequestContract, key string) []byte {
	if r.RequestWithJson() {
		val, _, _, _ := jsonparser.Get(r.GetContent(), key)

		return val
	}

	return nil
}

type Request struct {
	contracts.RequestContract
}

type MapInput struct {
	data map[string]interface{}
	raw  []byte
}

func (mi *MapInput) GetMap() map[string]interface{} {
	return mi.data
}

func (mi *MapInput) GetRaw() []byte {
	return mi.raw
}

func (mi *MapInput) ScanInput(data []byte) error {
	mi.raw = data
	return jsoniter.Unmarshal(mi.raw, &mi.data)
}

func (mi MapInput) MarshalJSON() ([]byte, error) {
	return mi.raw, nil
}

func (mi MapInput) Unmarshal(to interface{}) error {
	return jsoniter.Unmarshal(mi.raw, to)
}

func (mi MapInput) Get(key string, v interface{}) error {
	valRow, valType, _, e := jsonparser.Get(mi.raw, key)

	if e != nil {
		return e
	}

	if b, ok := v.(*[]byte); ok {
		*b = valRow
		return nil
	}

	switch valType {
	case jsonparser.String:
		if sp, ok := v.(*string); ok {
			*sp = string(valRow)
		}
	case jsonparser.Number:
		switch vp := v.(type) {
		case *float64:
			*vp, _ = strconv.ParseFloat(string(valRow), 64)
		case *float32:
			fl, _ := strconv.ParseFloat(string(valRow), 32)
			*vp = float32(fl)
		case *int64:
			*vp, _ = strconv.ParseInt(string(valRow), 10, 64)
		case *int:
			*vp, _ = strconv.Atoi(string(valRow))
		case *int32:
			i, _ := strconv.ParseInt(string(valRow), 10, 32)
			*vp = int32(i)
		case *int16:
			i, _ := strconv.ParseInt(string(valRow), 10, 16)
			*vp = int16(i)
		case *int8:
			i, _ := strconv.ParseInt(string(valRow), 10, 8)
			*vp = int8(i)
		case *uint64:
			*vp, _ = strconv.ParseUint(string(valRow), 10, 64)
		case *uint:
			ui, _ := strconv.ParseUint(string(valRow), 10, 64)
			*vp = uint(ui)
		case *uint32:
			ui, _ := strconv.ParseUint(string(valRow), 10, 32)
			*vp = uint32(ui)
		case *uint16:
			ui, _ := strconv.ParseUint(string(valRow), 10, 16)
			*vp = uint16(ui)
		case *uint8:
			ui, _ := strconv.ParseUint(string(valRow), 10, 8)
			*vp = uint8(ui)
		}

	case jsonparser.Boolean:
		if sp, ok := v.(*bool); ok {
			*sp, _ = strconv.ParseBool(string(valRow))
		}
	case jsonparser.Null, jsonparser.NotExist, jsonparser.Unknown:
		return nil
	default:
		return jsoniter.Unmarshal(valRow, v)
	}

	return nil
}

type JsonInputHandler func(j JsonInput)

type JsonInput []byte

func (j JsonInput) Get(key string) []byte {
	value, _, _, _ := jsonparser.Get(j, key)

	return value
}

func (j JsonInput) ParamBytes(key string) []byte {
	return nil
}

func (j JsonInput) File(key string) (contracts.UploadFile, error) {
	return nil, errors.New("jsonInput does not implement func File(key string) (content.UploadFile, error)")
}

func (j JsonInput) Each(h JsonInputHandler) error {
	_, e := jsonparser.ArrayEach(j, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		h(value)
	})

	return e
}

func (j JsonInput) GetValue(key ...string) contracts.InputValue {
	if len(key) > 0 {
		return j.Get(key[0])
	}

	return contracts.InputValue(j)
}

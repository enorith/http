package http

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/buger/jsonparser"
	"github.com/enorith/container"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	httpErrors "github.com/enorith/http/errors"
	"github.com/enorith/http/validation"
	"github.com/enorith/supports/byt"
	"github.com/enorith/supports/reflection"
)

var cs cacheStruct

var (
	typeRequest,
	typeParamInt64,
	typeParamString,
	typeParamInt,
	typeParamUnit reflect.Type
)

var (
	uploadFileType = reflect.TypeOf((*contracts.UploadFile)(nil)).Elem()
)

type jsonInputHandler func(j jsonInput)

type jsonInput []byte

func (j jsonInput) Get(key string) []byte {
	value, _, _, _ := jsonparser.Get(j, key)

	return value
}

func (j jsonInput) ParamBytes(key string) []byte {
	return nil
}

func (j jsonInput) File(key string) (contracts.UploadFile, error) {
	return nil, errors.New("jsonInput does not implement func File(key string) (content.UploadFile, error)")
}

func (j jsonInput) Each(h jsonInputHandler) error {
	_, e := jsonparser.ArrayEach(j, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		h(value)
	})

	return e
}

func (j jsonInput) GetValue(key ...string) contracts.InputValue {
	if len(key) > 0 {
		return j.Get(key[0])
	}

	return contracts.InputValue(j)
}

type cacheStruct struct {
	cache map[interface{}]bool
	mu    sync.RWMutex
}

func (c *cacheStruct) get(abs interface{}) (ok bool, exist bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ok, exist = c.cache[abs]

	return
}

func (c *cacheStruct) set(abs interface{}, b bool) {
	c.mu.Lock()
	c.cache[abs] = b
	c.mu.Unlock()
}

//RequestInjector inject request object, with validation
type RequestInjector struct {
	runtime      container.Interface
	request      contracts.RequestContract
	validator    *validation.Validator
	paramIndex   int
	requestIndex int
}

func (r *RequestInjector) Injection(abs interface{}, value reflect.Value) (reflect.Value, error) {
	var e error
	defer func() {
		if x := recover(); x != nil {
			value = reflect.Value{}
			if err, ok := x.(error); ok {
				e = err
			}
		}
	}()
	ts := reflection.StructType(abs)

	//value = last
	if r.isRequest(abs) {
		// dependency injection sub struct of content.Request
		tf := ts.Field(r.requestIndex).Type
		instanceReq, err := r.runtime.Instance(tf)
		if err != nil {
			return value, err
		}
		indVal := reflect.Indirect(value)

		indVal.Field(r.requestIndex).Set(instanceReq)

		e := r.unmarshal(indVal, r.request)
		if e != nil {
			return value, e
		}

		return value, nil
	} else if r.isParam(abs) {
		// parameter injection
		params := r.request.ParamsSlice()
		paramsLength := len(params)
		if paramsLength > r.paramIndex {
			param := params[r.paramIndex]
			if ts == typeParamInt64 || ts == typeParamInt {
				val, err := byt.ToInt64(param)
				if err != nil {

					return value, err
				}
				value.Elem().SetInt(val)
			} else if ts == typeParamUnit {

				val, err := byt.ToUint64(param)
				if err != nil {

					return value, err
				}
				value.Elem().SetUint(val)
			} else if ts == typeParamString {
				val := byt.ToString(param)

				value.Elem().SetString(val)
			}

			r.paramIndex++
		}

		return value, nil
	}

	return value, e
}

func (r *RequestInjector) When(abs interface{}) bool {
	ok, e := cs.get(abs)
	if e {
		return ok
	}

	// dependency is sub struct of content.Request
	is := r.isParam(abs) || r.isRequest(abs)
	cs.set(abs, is)

	return is
}

func (r *RequestInjector) isParam(abs interface{}) bool {
	ts := reflection.StructType(abs)

	return ts == typeParamInt || ts == typeParamString || ts == typeParamInt64 || ts == typeParamUnit
}

func (r *RequestInjector) isRequest(abs interface{}) bool {
	r.requestIndex = reflection.SubStructOf(abs, typeRequest)
	return r.requestIndex > -1
}

func (r *RequestInjector) unmarshal(value reflect.Value, request contracts.InputSource) error {
	typ := value.Type()
	if validated, ok := value.Interface().(validation.WithValidation); ok {
		rules := validated.Rules()
		var errs []string
		for attribute, rules := range rules {
			errs = append(errs, r.validator.PassesRules(request, attribute, rules)...)
		}
		if len(errs) > 0 {
			return httpErrors.UnprocessableEntity(errs[0])
		}
	}
	for i := 0; i < value.NumField(); i++ {
		f := value.Field(i)
		ft := typ.Field(i)
		if f.IsZero() {
			if input := ft.Tag.Get("input"); input != "" {
				e := r.passValidate(ft.Tag, request, input)
				if e != nil {
					return e
				}
				e = r.unmarshalField(f, request.Get(input))
				if e != nil {
					return e
				}
			} else if param := ft.Tag.Get("param"); param != "" {
				e := r.passValidate(ft.Tag, request, param)
				if e != nil {
					return e
				}
				e = r.unmarshalField(f, request.ParamBytes(param))
				if e != nil {
					return e
				}
			} else if file := ft.Tag.Get("file"); file != "" {
				e := r.passValidate(ft.Tag, request, file)
				if e != nil {
					return e
				}
				if f.Type() == uploadFileType {
					uploadFile, e := request.File(file)
					if e != nil {
						return httpErrors.UnprocessableEntity(
							fmt.Sprintf("attribute [%s] must be a file", file))
					}
					f.Set(reflect.ValueOf(uploadFile))
				}
			}
		}
	}
	return nil
}

func (r *RequestInjector) unmarshalField(field reflect.Value, data []byte) error {
	v := field.Interface()
	if _, ok := v.([]byte); ok {
		field.SetBytes(data)
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(byt.ToString(data))
	case reflect.Bool:
		in, _ := byt.ToBool(data)
		field.SetBool(in)
	case reflect.Int, reflect.Int32, reflect.Int64:
		in, _ := byt.ToInt64(data)
		field.SetInt(in)
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		in, _ := byt.ToUint64(data)
		field.SetUint(in)
	case reflect.Float32, reflect.Float64:
		in, _ := byt.ToFloat64(data)
		field.SetFloat(in)
	case reflect.Struct:
		newF := reflect.New(field.Type())
		newV := reflect.Indirect(newF)

		e := r.unmarshal(newV, jsonInput(data))
		if e != nil {
			return e
		}
		field.Set(newV)
	case reflect.Ptr:

		newF := reflect.New(field.Type().Elem())
		newV := reflect.Indirect(newF)

		e := r.unmarshal(newV, jsonInput(data))
		if e != nil {
			return e
		}
		field.Set(newF)
	case reflect.Slice:
		it := field.Type().Elem()
		var ivs []reflect.Value

		jsonInput(data).Each(func(j jsonInput) {
			itv := reflect.New(it)
			itve := reflect.Indirect(itv)
			r.unmarshalField(itve, j)
			ivs = append(ivs, itve)
		})
		l := len(ivs)
		slice := reflect.MakeSlice(field.Type(), l, l)
		for index, v := range ivs {
			slice.Index(index).Set(v)
		}

		field.Set(slice)
	}

	return nil
}

func (r *RequestInjector) passValidate(tag reflect.StructTag, request contracts.InputSource, attribute string) error {
	if rule := tag.Get("validate"); rule != "" {
		rules := strings.Split(rule, "|")

		errs := r.validator.Passes(request, attribute, rules)
		if len(errs) > 0 {
			return httpErrors.UnprocessableEntity(errs[0])
		}
	}
	return nil
}

func init() {
	typeParamInt64 = reflection.StructType(content.ParamInt64(42))
	typeParamString = reflection.StructType(content.Param("42"))
	typeParamUnit = reflection.StructType(content.ParamUint64(42))
	typeParamInt = reflection.StructType(content.ParamInt(42))
	typeRequest = reflection.StructType(content.Request{})
	cs = cacheStruct{cache: map[interface{}]bool{}, mu: sync.RWMutex{}}
}

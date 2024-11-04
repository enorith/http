package http

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/enorith/container"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	httpErrors "github.com/enorith/http/errors"
	"github.com/enorith/http/validation"
	"github.com/enorith/supports/byt"
	"github.com/enorith/supports/reflection"
	jsoniter "github.com/json-iterator/go"
)

var cs cacheStruct

var (
	typeRequest,
	typeJsonRequest,
	typeParamInt64,
	typeParamString,
	typeParamInt,
	typeParamUnit reflect.Type
)

var (
	uploadFileType = reflect.TypeOf((*contracts.UploadFile)(nil)).Elem()
)

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

// RequestInjector inject request object, with validation
type RequestInjector struct {
	runtime    container.Interface
	request    contracts.RequestContract
	validator  *validation.Validator
	paramIndex int
}

func (r *RequestInjector) Injection(abs interface{}, value reflect.Value) (reflect.Value, error) {
	var e error
	defer func() {
		if x := recover(); x != nil {
			fmt.Println("[request injection] panic: \n" + string(debug.Stack()))
			value = reflect.Value{}
			if err, ok := x.(error); ok {
				e = err
			}

			if err, ok := x.(string); ok {
				e = errors.New(err)
			}
		}
	}()
	ts := reflection.StructType(abs)

	reqIndex := r.reqIndex(abs)
	jsonReqIndex := r.jsonReqIndex(abs)

	//value = last
	if reqIndex > -1 || jsonReqIndex > -1 {
		indVal := reflect.Indirect(value)

		if reqIndex > -1 {
			// dependency injection sub struct of content.Request
			tf := ts.Field(reqIndex).Type
			instanceReq, err := r.runtime.Instance(tf)
			if err != nil {
				return value, err
			}

			indVal.Field(reqIndex).Set(instanceReq)
		} else if jsonReqIndex > -1 {
			tf := ts.Field(jsonReqIndex).Type
			instanceReq, err := r.runtime.Instance(tf)
			if err != nil {
				return value, err
			}

			indVal.Field(jsonReqIndex).Set(instanceReq)
		}

		if jsonReqIndex > -1 {

			e := r.request.Unmarshal(value.Interface())

			if e != nil {
				return value, e
			}

			validateError := r.validate(value, r.request)

			if len(validateError) > 0 {
				return value, validateError
			}

			return value, e
		}

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
	is := r.isParam(abs) || r.reqIndex(abs) > -1 || r.jsonReqIndex(abs) > -1
	cs.set(abs, is)

	return is
}

func (r *RequestInjector) isParam(abs interface{}) bool {
	ts := reflection.StructType(abs)

	return ts == typeParamInt || ts == typeParamString || ts == typeParamInt64 || ts == typeParamUnit
}

func (r *RequestInjector) reqIndex(abs interface{}) int {
	return reflection.SubStructOf(abs, typeRequest)
}

func (r *RequestInjector) jsonReqIndex(abs interface{}) int {
	return reflection.SubStructOf(abs, typeJsonRequest)
}

func (r *RequestInjector) unmarshal(value reflect.Value, request contracts.InputSource) error {
	typ := value.Type()
	validateError := make(validation.ValidateError)

	for i := 0; i < value.NumField(); i++ {
		f := value.Field(i)
		ft := typ.Field(i)
		if f.IsZero() {
			input := ft.Tag.Get("input")
			if input == "" {
				input = ft.Tag.Get("json")
			}

			if input != "" && input != "-" {
				errs := r.passValidate(ft.Tag, request, input)
				if len(errs) > 0 {
					validateError[input] = errs
					continue
				}
				e := r.unmarshalField(f, request.Get(input))
				if e != nil {
					return fmt.Errorf("[request injection] unmarshal request field \"%s\" error, check your type definition: %s", input, e.Error())
				}
			} else if param := ft.Tag.Get("param"); param != "" {
				errs := r.passValidate(ft.Tag, request, param)
				if len(errs) > 0 {
					validateError[param] = errs
					continue
				}
				if rc, ok := request.(contracts.RequestContract); ok {
					e := r.unmarshalField(f, rc.ParamBytes(param))
					if e != nil {
						return e
					}
				}
			} else if file := ft.Tag.Get("file"); file != "" {
				errs := r.passValidate(ft.Tag, request, file)
				if len(errs) > 0 {
					validateError[file] = errs
					continue
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
			if ft.Anonymous && f.Type() != typeRequest {
				e := r.unmarshal(f, request)
				if e != nil {
					return e
				}
			}
		}
	}

	ve := r.validate(value, request)

	for k, v := range ve {
		validateError[k] = v
	}

	if len(validateError) > 0 {
		return validateError
	}

	return nil
}

func (r *RequestInjector) validate(value reflect.Value, request contracts.InputSource) validation.ValidateError {
	validateError := make(validation.ValidateError)

	if validated, ok := value.Interface().(validation.WithValidation); ok {
		rules := validated.Rules()
		for attribute, rules := range rules {
			errs := r.validator.PassesRules(request, attribute, rules)
			if len(errs) > 0 {
				validateError[attribute] = errs
			}
		}
	}

	return validateError
}

func (r *RequestInjector) unmarshalField(field reflect.Value, data []byte) error {
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		return nil
	}

	v := field.Interface()
	if _, ok := v.([]byte); ok {
		field.SetBytes(data)
		return nil
	}
	newF := reflect.New(field.Type())

	newInterface := newF.Interface()
	if fv, ok := newInterface.(contracts.InputScanner); ok && len(data) > 0 {
		e := fv.ScanInput(data)
		if e == nil {
			field.Set(newF.Elem())
		}

		return e
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(byt.ToString(data))
	case reflect.Bool:
		in, _ := byt.ToBool(data)
		field.SetBool(in)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		in, _ := byt.ToInt64(data)
		field.SetInt(in)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		in, _ := byt.ToUint64(data)
		field.SetUint(in)
	case reflect.Float32, reflect.Float64:
		in, _ := byt.ToFloat64(data)
		field.SetFloat(in)
	case reflect.Map:
		newM := reflect.New(field.Type())

		e := jsoniter.Unmarshal(data, newM.Interface())
		if e != nil {
			return e
		}
		field.Set(reflect.Indirect(newM))
	case reflect.Struct:
		newF := reflect.New(field.Type())
		newV := reflect.Indirect(newF)

		e := r.unmarshal(newV, content.JsonInput(data))
		if e != nil {
			return e
		}
		field.Set(newV)
	case reflect.Ptr:
		newF := reflect.New(field.Type().Elem())
		newV := reflect.Indirect(newF)

		e := r.unmarshal(newV, content.JsonInput(data))
		if e != nil {
			return e
		}
		field.Set(newF)
	case reflect.Slice:
		it := field.Type().Elem()
		if it.Kind() == reflect.Uint8 {
			// []uint8 as []byte
			field.SetBytes(data)
			return nil
		}

		var ivs []reflect.Value

		if e := content.JsonInput(data).Each(func(j content.JsonInput) error {
			itv := reflect.New(it)
			itve := reflect.Indirect(itv)
			if e := r.unmarshalField(itve, j); e != nil {
				return e
			}
			ivs = append(ivs, itve)

			return nil
		}); e != nil {

			return e
		}

		l := len(ivs)
		slice := reflect.MakeSlice(field.Type(), l, l)
		for index, v := range ivs {
			slice.Index(index).Set(v)
		}

		field.Set(slice)
	}

	return nil
}

func (r *RequestInjector) passValidate(tag reflect.StructTag, request contracts.InputSource, attribute string) []string {
	if rule := tag.Get("validate"); rule != "" {
		rules := strings.Split(rule, "|")

		return r.validator.Passes(request, attribute, rules)
	}
	return nil
}

func init() {
	typeParamInt64 = reflection.StructType(content.ParamInt64(42))
	typeParamString = reflection.StructType(content.Param("42"))
	typeParamUnit = reflection.StructType(content.ParamUint64(42))
	typeParamInt = reflection.StructType(content.ParamInt(42))
	typeRequest = reflection.StructType(content.Request{})
	typeJsonRequest = reflection.StructType(content.JsonRequest{})
	cs = cacheStruct{cache: map[interface{}]bool{}, mu: sync.RWMutex{}}
}

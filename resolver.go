package http

import (
	"context"
	"reflect"

	"github.com/enorith/container"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/validation"
	"github.com/enorith/supports/reflection"
)

var (
	contextType = reflection.InterfaceType[context.Context]()
)

type KernelRequestResolver struct {
}

func (rr KernelRequestResolver) ResolveRequest(r contracts.RequestContract, runtime container.Interface) {
	runtime.RegisterSingleton(r)

	runtime.Singleton(reflect.TypeOf((*contracts.RequestContract)(nil)).Elem(), r)

	runtime.BindFunc(&content.Request{}, func(c container.Interface) (interface{}, error) {

		return &content.Request{RequestContract: r}, nil
	}, true)

	runtime.BindFunc(content.Request{}, func(c container.Interface) (interface{}, error) {

		return content.Request{RequestContract: r}, nil
	}, true)

	runtime.WithInjector(&RequestInjector{runtime: runtime, request: r, validator: validation.DefaultValidator})

	runtime.BindFunc(contextType, func(c container.Interface) (interface{}, error) {
		return r.Context(), nil
	}, true)
}

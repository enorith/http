# Http component for [Enorith](https://github.com/enorith/framework)

## Usage

### Basic example

```golang
package main

import (
	"github.com/enorith/container"
	"github.com/enorith/http"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/router"
	"log"
	"fmt"
)

type FooRequest struct {
	content.Request
    // input injection and validation
	Foo string `input:"foo" validate:"required"`
	File contracts.UploadFile `file:"file"`
}

func main() {
	s := http.NewServer(func() *container.Container {
		return container.New() // runtime IoC container
	}, true)
	
	e := s.Serve(":10800", func(ro *router.Wrapper, k *http.Kernel) {
		//k.OutputLog = true
		ro.HandleGet("/", func(r contracts.RequestContract) contracts.ResponseContract {

			return content.TextResponse("ok", 200)
		})

		// request injection
		ro.Get("/foo", func(r FooRequest) contracts.ResponseContract {

			return content.TextResponse("input foo: " + r.Foo, 200)
		})

		// param injection
		ro.Get("/:id", func(id content.ParamUint64) string {

			return fmt.Sprintf("input id: %d", id)
		})
	})

	if e != nil {
		log.Fatalf("serve error: %v", e)
	}
}
```
### Middleware

```golang
package foo

import (
	"github.com/enorith/http"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/router"
)

type FooMiddleware struct {

}

func (FooMiddleware) Handle(r contracts.RequestContract, next http.PipeHandler) contracts.ResponseContract {
	//TODO: before request handled

	resp := next(r)

	//TOD: after request handled
	return resp
}


func Handler(ro *router.Wrapper, k *http.Kernel) {
	k.SetMiddleware([]http.RequestMiddleware{ // global middleware
		FooMiddleware{},
	})

	k.SetMiddlewareGroup(map[string][]http.RequestMiddleware{
		"foo.mid": {FooMiddleware{}}
	})

	ro.Get("foo", fun() string { return "bar"}).Middleware("foo.mid")
}
```
### With Container

```golang
package main

import (
	"reflect"

	"github.com/enorith/container"
	"github.com/enorith/http"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/router"
)

type FooStruct struct {
	Bar string
}

func main() {
	srv := http.NewServer(func(request contracts.RequestContract) container.Interface {
		con := container.New()

		con.BindFunc(FooStruct{}, func(c container.Interface) (reflect.Value, error) { // bind instance
			return reflect.ValueOf(FooStruct{Bar: "baz"}), nil
		}, false)
		return con
	}, true)

	srv.Serve(":8000", func(rw *router.Wrapper, k *http.Kernel) {


		rw.Get("foo", func(fs FooStruct) string { // parameter injection
			return fs.Bar
		})
	})
}

```

## TODO

- [ ] Get client ip behand proxy
- [ ] Validation (incomplete)
- [ ] Logging

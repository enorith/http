# Http component for [Enorith](https://github.com/enorith/framework)

## Basic usage

```go
package main

import (
	"github.com/enorith/container"
	"github.com/enorith/http"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/router"
	"log"
)

type FooRequest struct {
	content.Request
    // input injection and validation
	Foo string `input:"foo" validate:"required"`
}

func main() {
	s := http.NewServer(func() *container.Container {
		return container.New()
	}, true)
	
	e := s.Serve(":10800", func(ro *router.Wrapper, k *http.Kernel) {
		//k.OutputLog = true
		ro.HandleGet("/", func(r contracts.RequestContract) contracts.ResponseContract {

			return content.TextResponse("ok", 200)
		})

		ro.Get("/foo", func(r FooRequest) contracts.ResponseContract {

			return content.TextResponse("input foo: " + r.Foo, 200)
		})
	})

	if e != nil {
		log.Fatalf("serve error: %v", e)
	}
}
```
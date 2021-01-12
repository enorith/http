package router

import (
	"net/http"

	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
)

func NetHttpHandlerFromHttp(request *content.NetHttpRequest, h http.Handler) contracts.ResponseContract {
	r, ow := request.Origin(), request.OriginWriter()

	h.ServeHTTP(ow, r)

	return content.NewHandledResponse()
}

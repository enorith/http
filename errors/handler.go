package errors

import (
	"github.com/enorith/exception"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
)

type ErrorHandler interface {
	HandleError(e interface{}, r contracts.RequestContract) contracts.ResponseContract
}

type StandardErrorHandler struct {
	Debug bool
}

func (h *StandardErrorHandler) HandleError(e interface{}, r contracts.RequestContract) contracts.ResponseContract {
	return h.BaseHandle(e, r)
}

func (h *StandardErrorHandler) BaseHandle(e interface{}, r contracts.RequestContract) contracts.ResponseContract {
	var ex exception.Exception
	var code = 500
	var headers map[string]string = nil
	if t, ok := e.(string); ok {
		ex = exception.NewException(t, code)
	} else if t, ok := e.(exception.HttpException); ok {
		ex = t
		headers = t.Headers()
	} else if t, ok := e.(exception.Exception); ok {
		ex = t
	} else if t, ok := e.(error); ok {
		ex = exception.NewExceptionFromError(t, code)
	} else {
		ex = exception.NewException("undefined exception", code)
	}

	if t, ok := e.(contracts.WithStatusCode); ok {
		code = t.StatusCode()
	}

	if r.ExceptsJson() {
		return content.JsonErrorResponseFormatter(ex, code, h.Debug, headers)
	} else {
		//tmp := fmt.Sprintf("%s/errors/%d.html", h.App.Structure().BasePath, code)
		//fmt.Println(tmp)
		//if e, _ := file.PathExists(tmp); e {
		//	return content.FileResponse(tmp, 200, ex)
		//}
		return content.HtmlErrorResponseFormatter(ex, code, h.Debug, headers)
	}
}

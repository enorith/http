package errors

import (
	"fmt"
	"html/template"

	"github.com/enorith/exception"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/errors/assets"
	"github.com/enorith/supports/file"
)

type Trace struct {
	File string
	Line int
}

type ErrorData struct {
	Code      int
	File      string
	Line      int
	Message   string
	Debug     bool
	Recovered bool
	Fatal     bool
	Traces    []Trace
}

type ErrorHandler interface {
	HandleError(e interface{}, r contracts.RequestContract, recovered bool) contracts.ResponseContract
}

type StandardErrorHandler struct {
	Debug bool
}

func (h *StandardErrorHandler) HandleError(e interface{}, r contracts.RequestContract, recovered bool) contracts.ResponseContract {
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
		return JsonErrorResponseFormatter(ex, code, h.Debug, recovered, headers)
	} else {
		//tmp := fmt.Sprintf("%s/errors/%d.html", h.App.Structure().BasePath, code)
		//fmt.Println(tmp)
		//if e, _ := file.PathExists(tmp); e {
		//	return content.FileResponse(tmp, 200, ex)
		//}
		te := fmt.Sprintf("%d.html", code)
		if !file.PathExistsFS(assets.FS, te) {
			te = "error.html"
		}

		temp, _ := template.ParseFS(assets.FS, te)

		return content.TempResponse(temp, code, toErrorData(code, ex, h.Debug, recovered))
		// return content.HtmlErrorResponseFormatter(ex, code, h.Debug, headers)
	}
}

func JsonErrorResponseFormatter(err exception.Exception, code int, debug, recovered bool, headers map[string]string) contracts.ResponseContract {

	data := map[string]interface{}{
		"code":    code,
		"message": err.Error(),
	}

	if debug {
		data["file"] = err.File()
		data["line"] = err.Line()
		data["recoverd"] = recovered
		data["traces"] = func() []string {
			var traces []string
			for k, v := range err.Traces() {
				frame := fmt.Sprintf("#%d %s [%d]: %s", k+1, v.File(), v.Line(), v.Name())
				traces = append(traces, frame)
			}
			return traces
		}()
	}

	return content.JsonResponse(data, code, headers)
}

func toErrorData(code int, err exception.Exception, debug, recoverd bool) (data ErrorData) {

	data.Message = err.Error()
	data.File = err.File()
	data.Code = code
	data.Debug = debug
	for _, t := range err.Traces() {
		data.Traces = append(data.Traces, Trace{
			File: t.File(),
			Line: t.Line(),
		})
	}
	data.Recovered = recoverd
	data.Fatal = code >= 500
	return
}

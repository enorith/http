package errors

import (
	"fmt"
	"html/template"

	"github.com/enorith/exception"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/errors/assets"
	"github.com/enorith/http/validation"
	"github.com/enorith/http/view"
	"github.com/enorith/supports/file"
)

type Trace struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

type ErrorData struct {
	Code      int                      `json:"code"`
	File      string                   `json:"file,omitempty"`
	Line      int                      `json:"line,omitempty"`
	Message   string                   `json:"message"`
	Debug     bool                     `json:"debug,omitempty"`
	Recovered bool                     `json:"recovered,omitempty"`
	Fatal     bool                     `json:"fatal,omitempty"`
	Traces    []Trace                  `json:"traces,omitempty"`
	Errors    validation.ValidateError `json:"errors,omitempty"`
}

type ErrorHandler interface {
	HandleError(e interface{}, r contracts.RequestContract, recovered bool) contracts.ResponseContract
}

type StandardErrorHandler struct {
	Debug bool
}

func (h *StandardErrorHandler) HandleError(e interface{}, r contracts.RequestContract, recovered bool) contracts.ResponseContract {

	var headers map[string]string
	if t, ok := e.(exception.HttpException); ok {
		headers = t.Headers()
	}

	errorData := ParseError(e, h.Debug, recovered)
	code := errorData.Code
	if r.ExceptsJson() {
		return content.JsonResponse(errorData, code, headers)
	} else {
		if v, e := view.View(fmt.Sprintf("errors.%d", code), code, errorData); e == nil {
			return v
		}

		te := fmt.Sprintf("%d.html", code)
		if !file.PathExistsFS(assets.FS, te) {
			te = "error.html"
		}

		temp, _ := template.ParseFS(assets.FS, te)

		return content.TempResponse(temp, code, errorData)
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
	data.Code = code
	data.Debug = debug
	if debug {
		data.File = err.File()
		data.Line = err.Line()

		for _, t := range err.Traces() {
			data.Traces = append(data.Traces, Trace{
				File: t.File(),
				Line: t.Line(),
			})
		}
	}

	e := err.GetError()

	if ve, ok := e.(validation.ValidateError); ok {
		data.Errors = ve
	}

	data.Recovered = recoverd
	data.Fatal = code >= 500
	return
}

func ParseError(e interface{}, debug, recovered bool) ErrorData {
	var ex exception.Exception
	var code = 500
	if t, ok := e.(string); ok {
		ex = exception.NewException(t, code)
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
	return toErrorData(code, ex, debug, recovered)
}

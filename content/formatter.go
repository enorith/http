package content

import (
	"fmt"
	"github.com/enorith/exception"
	"github.com/enorith/http/contracts"
)

type ErrorResponseFormatter func(err exception.Exception, code int, debug bool, headers map[string]string) contracts.ResponseContract

func JsonErrorResponseFormatter(err exception.Exception, code int, debug bool, headers map[string]string) contracts.ResponseContract {

	data := map[string]interface{}{
		"code":    err.Code(),
		"message": err.Error(),
	}

	if debug {
		data["file"] = err.File()
		data["line"] = err.Line()
		data["traces"] = func() []string {
			var traces []string
			for k, v := range err.Traces() {
				frame := fmt.Sprintf("#%d %s [%d]: %s", k+1, v.File(), v.Line(), v.Name())
				traces = append(traces, frame)
			}
			return traces
		}()
	}

	return JsonResponse(data, code, headers)
}

func HtmlErrorResponseFormatter(err exception.Exception, code int, debug bool, headers map[string]string) contracts.ResponseContract {
	html := `
<!DOCTYPE html>
<html>
    <head>
        <meta charset="utf-8" />
        <meta name="robots" content="noindex,nofollow" />
        <style>%s</style>
    </head>
    <body>
        <div id="content">
			<div class="message">%s</div>
        	<ul class="trace">%s</div>
		</div>
    </body>
</html>
`
	if headers == nil {
		headers = map[string]string{}
	}

	css := `
.message {
	width: 100%;
	font-size: 28px;
	color: red;
	margin: 20px 0;
	text-align: center;
}
.trace {
	list-style: none;
	width: 80%;
	margin-left: auto;
	margin-right: auto;
}
#content {
	width: 80%;
	display: block;
	margin: auto;
}
`

	headers["Content-type"] = ContentTypeHtml

	return NewResponse([]byte(fmt.Sprintf(html, css, err.Error(), func() string {
		if !debug {
			return ""
		}
		ts := ""
		for k, v := range err.Traces() {
			frame := fmt.Sprintf("#%d %s [%d]: %s", k+1, v.File(), v.Line(), v.Name())
			ts += fmt.Sprintf("<li>%s</li>", frame)
		}
		return ts
	}())), headers, code)
}

package router

import (
	"bufio"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	"github.com/valyala/fasthttp"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type netHTTPBody struct {
	b []byte
}

func (r *netHTTPBody) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.b)
	r.b = r.b[n:]
	return n, nil
}

func (r *netHTTPBody) Close() error {
	r.b = r.b[:0]
	return nil
}

type netHTTPResponseWriter struct {
	statusCode int
	h          http.Header
	body       []byte
	ctx        *fasthttp.RequestCtx
}

func (w *netHTTPResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn := w.ctx.Conn()

	return conn, bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}

func (w *netHTTPResponseWriter) StatusCode() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}
	return w.statusCode
}

func (w *netHTTPResponseWriter) Header() http.Header {
	if w.h == nil {
		w.h = make(http.Header)
	}
	return w.h
}

func (w *netHTTPResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *netHTTPResponseWriter) Write(p []byte) (int, error) {
	w.body = append(w.body, p...)
	return len(p), nil
}

func FastHttpHandlerFromHttp(req *content.FastHttpRequest, h http.Handler) contracts.ResponseContract {
	ctx := req.Origin()
	var r http.Request
	body := ctx.PostBody()
	r.Method = string(ctx.Method())
	if ctx.Request.Header.IsHTTP11() {
		r.Proto = "HTTP/1.1"
	} else {
		r.Proto = "HTTP/1.0"
	}
	r.URL = getUrl(ctx)
	r.ProtoMajor = 1
	r.ProtoMinor = 1
	r.RequestURI = string(ctx.RequestURI())
	r.ContentLength = int64(len(body))
	r.Host = string(ctx.Host())
	r.RemoteAddr = ctx.RemoteAddr().String()

	hdr := make(http.Header)
	ctx.Request.Header.VisitAll(func(k, v []byte) {
		sk := string(k)
		sv := string(v)
		switch sk {
		case "Transfer-Encoding":
			r.TransferEncoding = append(r.TransferEncoding, sv)
		default:
			hdr.Set(sk, sv)
		}
	})
	r.Header = hdr
	r.Body = &netHTTPBody{body}
	rURL, err := url.ParseRequestURI(r.RequestURI)
	if err != nil {
		return content.TextResponse("Internal Server Error", fasthttp.StatusInternalServerError)
	}
	r.URL = rURL

	w := netHTTPResponseWriter{ctx: ctx}
	h.ServeHTTP(&w, &r)

	response := content.NewResponse(w.body, nil, w.StatusCode())
	for k, vv := range w.Header() {
		v := strings.Join(vv, "; ")
		response.SetHeader(k, v)
	}
	return response
}

func getUrl(r *fasthttp.RequestCtx) *url.URL {
	uri := r.URI()
	u, _ := url.Parse(uri.String())
	return u
}

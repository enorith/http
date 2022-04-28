package contracts

import (
	"html/template"
	"net/http"
)

type ResponseContract interface {
	Content() []byte
	Headers() map[string]string
	SetHeader(key string, value string) ResponseContract
	SetHeaders(headers map[string]string) ResponseContract
	StatusCode() int
	SetStatusCode(code int) ResponseContract
	Handled() bool
}

type TemplateResponseContract interface {
	Template() *template.Template
	TemplateData() interface{}
}

type WithStatusCode interface {
	StatusCode() int
}

type WithResponseCode interface {
	ResponseCode() int
}

type WithContentType interface {
	ContentType() string
}

type HTMLer interface {
	HTML() []byte
}

type WithHeaders interface {
	Headers() map[string]string
}

type FileServer interface {
	Path() string
	Prefix() string
}

type WithResponseCookies interface {
	Cookies() []*http.Cookie
	SetCookie(cookie *http.Cookie)
	ClearCookies()
}

package contracts

import (
	"html/template"
)

type ResponseContract interface {
	Content() []byte
	Headers() map[string]string
	SetHeader(key string, value string) ResponseContract
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

type WithHeaders interface {
	Headers() map[string]string
}

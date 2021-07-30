package errors

import "net/http"

type UnprocessableEntity string

func (u UnprocessableEntity) StatusCode() int {
	return 422
}

func (u UnprocessableEntity) Error() string {
	return string(u)
}

type AccessDenied string

func (ad AccessDenied) Error() string {
	return string(ad)
}

func (ad AccessDenied) StatusCode() int {
	return 403
}

type Unauthorized string

func (u Unauthorized) Error() string {
	return string(u)
}

func (u Unauthorized) StatusCode() int {
	return 401
}

type BadRequest string

func (br BadRequest) Error() string {
	return string(br)
}

func (br BadRequest) StatusCode() int {
	return 400
}

type NotFound string

func (nf NotFound) Error() string {
	return string(nf)
}

func (nf NotFound) StatusCode() int {
	return 404
}

type Code int

func (c Code) Error() string {
	return http.StatusText(int(c))
}

func (c Code) StatusCode() int {
	return int(c)
}

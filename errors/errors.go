package errors

type UnprocessableEntityError string

func (u UnprocessableEntityError) StatusCode() int {
	return 422
}

func (u UnprocessableEntityError) Error() string {
	return string(u)
}
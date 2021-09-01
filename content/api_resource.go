package content

import (
	jsoniter "github.com/json-iterator/go"
)

var (
	DefaultDataKey    = "data"
	DefaultCodeKey    = "code"
	DefaultMessageKey = "message"
	DefaultCode       = 200
)

type ApiResouce struct {
	dataKey    string
	data       interface{}
	codeKey    string
	code       int
	statusCode int
	messageKey string
	message    string
}

func (ar *ApiResouce) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j[ar.codeKey] = ar.code
	j[ar.dataKey] = ar.data
	j[ar.messageKey] = ar.message
	return jsoniter.Marshal(j)
}

func (ar *ApiResouce) Code(code int) *ApiResouce {
	ar.code = code
	return ar
}

func (ar *ApiResouce) WithStatus(statusCode int) *ApiResouce {
	ar.statusCode = statusCode
	return ar
}

func (ar *ApiResouce) Data(data interface{}) *ApiResouce {
	ar.data = data
	return ar
}

func (ar *ApiResouce) Message(message string) *ApiResouce {
	ar.message = message
	return ar
}

func (ar *ApiResouce) StatusCode() int {
	if ar.statusCode == 0 {
		return 200
	}

	return ar.statusCode
}

func ResourceResponse(data interface{}) *ApiResouce {
	return &ApiResouce{
		dataKey:    DefaultDataKey,
		codeKey:    DefaultCodeKey,
		messageKey: DefaultMessageKey,
		data:       data,
		code:       DefaultCode,
	}
}

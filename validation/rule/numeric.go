package rule

import "github.com/enorith/http/contracts"

type NumberType string

const (
	TypeInterger NumberType = "integer"
	TypeFloat    NumberType = "float"
	TypeInt      NumberType = "int"
)

type NumericInput struct {
	t NumberType
}

func (ni NumericInput) Passes(input contracts.InputValue) (success bool, skipAll bool) {
	switch ni.t {
	case TypeInterger, TypeInt:
		_, err := input.GetInt()
		if err == nil {
			return true, false
		}
	case TypeFloat:
		_, err := input.GetFloat()
		if err == nil {
			return true, false
		}
	}

	return false, false
}

func Numeric(t NumberType) NumericInput {
	return NumericInput{t}
}

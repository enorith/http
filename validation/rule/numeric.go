package rule

import "github.com/enorith/http/contracts"

type NumberType string

const (
	TypeInterger NumberType = "integer"
	TypeFloat    NumberType = "float"
)

type NumericInput struct {
	t NumberType
}

func (ni NumericInput) Passes(input contracts.InputValue) (success bool, skipAll bool) {

	return true, false
}

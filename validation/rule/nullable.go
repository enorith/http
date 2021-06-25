package rule

import "github.com/enorith/http/contracts"

type NullableInput struct {
}

func (ni NullableInput) Passes(input contracts.InputValue) (success bool, skipAll bool) {
	if len(input) == 0 {
		return true, true
	}

	return true, false
}

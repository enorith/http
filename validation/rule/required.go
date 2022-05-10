package rule

import (
	"github.com/enorith/http/contracts"
)

type Required struct {
}

func (r Required) Passes(input contracts.InputValue) (success bool, skipAll bool) {
	if len(input) > 0 {
		return true, false
	}

	return false, false
}

type RequiredIfRule struct {
	fn func() bool
}

func (ri RequiredIfRule) Passes(input contracts.InputValue) (success bool, skipAll bool) {
	i := ri.fn()
	if i {
		if len(input) > 0 {
			return true, false
		}
	}

	return !i, false
}

func RequiredIf(fn func() bool) RequiredIfRule {
	return RequiredIfRule{fn}
}

package rule

import (
	"github.com/enorith/http/contracts"
)

type Required struct {
	Attribute string
	Source    contracts.InputSource
}

func (r Required) Passes(input contracts.InputValue) (success bool, skipAll bool) {
	if len(input) > 0 {
		return true, false
	}

	f, _ := r.Source.File(r.Attribute)

	if f != nil {
		return true, false
	}

	return false, false
}

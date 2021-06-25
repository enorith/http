package rule

import (
	"github.com/enorith/http/contracts"
)

type FileInput struct {
	Attribute string
	Source    contracts.InputSource
}

func (f FileInput) Passes(input contracts.InputValue) (success bool, skipAll bool) {
	file, _ := f.Source.File(f.Attribute)

	if file != nil {
		return true, false
	}

	return false, false
}

package rule

import (
	"github.com/enorith/http/contracts"
	"github.com/enorith/language"
)

type Required struct {
	Attribute string
}

func (r Required) Passes(input contracts.InputValue) (success bool, skipAll bool) {
	if len(input) > 0 {
		return true, false
	}

	return false, false
}

func (r Required) Message() string {
	s, _ := language.T("validation", "required", map[string]string{
		"attribute": r.Attribute,
	})

	return s
}

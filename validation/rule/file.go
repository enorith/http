package rule

import (
	"github.com/enorith/http/contracts"
	"github.com/enorith/language"
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
func (f FileInput) Message() string {
	s, _ := language.T("validation", "file", map[string]string{
		"attribute": f.Attribute,
	})

	return s
}

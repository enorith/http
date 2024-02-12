package rule

import (
	"github.com/enorith/http/contracts"
	"github.com/enorith/supports/collection"
)

type InRule struct {
	values []string
}

func (i InRule) Passes(input contracts.InputValue) (success bool, skipAll bool) {
	return collection.IndexOf(i.values, string(input)) > -1, false
}

func In(values ...string) InRule {
	return InRule{values}
}

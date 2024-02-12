package rule

import (
	"github.com/enorith/http/contracts"
	"github.com/enorith/supports/carbon"
)

type DatetimeInput struct {
	formats []string
}

func (d DatetimeInput) Passes(input contracts.InputValue) (success bool, skipAll bool) {
	if len(input) < 1 {
		return true, false
	}

	_, e := carbon.Parse(string(input), nil, d.formats...)
	if e == nil {
		return true, false
	}

	return false, false
}

func Datetime(formats ...string) DatetimeInput {
	return DatetimeInput{formats}
}

package rule

import "github.com/enorith/http/contracts"

type Rule interface {
	Passes(input contracts.InputValue) (success bool, skipAll bool)
	Message() string
}

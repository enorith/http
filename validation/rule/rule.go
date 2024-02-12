package rule

import "github.com/enorith/http/contracts"

type Rule interface {
	Passes(input contracts.InputValue) (success bool, skipAll bool)
}

type Messager interface {
	Message() string
}

type Namer interface {
	RoleName() string
}

package validation

import (
	"fmt"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/validation/rule"
	"strings"
	"sync"
)

type WithValidation interface {
	Rules() map[string][]interface{}
}

var DefaultValidator *Validator

type ValidateError map[string]string

func (v ValidateError) StatusCode() int {
	return 422
}

func (v ValidateError) Error() string {
	var first string

	for _, s := range v {
		first = s
		break
	}

	return first
}

func Register(name string, register RuleRegister) {
	DefaultValidator.Register(name, register)
}

type RuleRegister func(attribute string, r contracts.InputSource, args ...string) rule.Rule

type Validator struct {
	registers map[string]RuleRegister
	mu        sync.RWMutex
}

func (v *Validator) Register(name string, register RuleRegister) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.registers[name] = register
}

func (v *Validator) GetRule(name string) (RuleRegister, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	r, ok := v.registers[name]

	return r, ok
}

func (v *Validator) Passes(req contracts.InputSource, attribute string, rules []string) (errors []string) {
	input := req.GetValue(attribute)
	for _, s := range rules {
		ss := strings.Split(s, ":")
		var args []string
		if len(ss) > 1 {
			args = strings.Split(ss[1], ",")
		}
		r, exist := v.GetRule(ss[0])
		if exist {
			inputRule := r(attribute, req, args...)
			success, skip := inputRule.Passes(input)
			if !success {
				message := inputRule.Message()

				if len(message) < 1 {
					message =  fmt.Sprintf("validation attribute [%s] error, rule [%s]", attribute, ss[0])
				}

				errors = append(errors, message)
			}

			if skip {
				break
			}
		}
	}
	return
}

func (v *Validator) PassesRules(req contracts.InputSource, attribute string, rules []interface{}) (errors []string) {
	input := req.GetValue(attribute)
	for _, rl := range rules {
		if s, ok := rl.(string); ok {
			ss := strings.Split(s, ":")
			var args []string
			if len(ss) > 1 {
				args = strings.Split(ss[1], ",")
			}
			r, exist := v.GetRule(ss[0])
			if exist {
				inputRule := r(attribute, req, args...)
				success, skip := inputRule.Passes(input)
				if !success {
					errors = append(errors, inputRule.Message())
				}

				if skip {
					break
				}
			}
		} else if rr, ok := rl.(rule.Rule); ok {
			success, skip := rr.Passes(input)
			if skip {
				break
			}

			if !success {
				errors = append(errors, rr.Message())
			}
		}
	}

	return
}

func init() {
	DefaultValidator = &Validator{registers: map[string]RuleRegister{}, mu: sync.RWMutex{}}
	Register("required", func(attribute string,r contracts.InputSource, args ...string) rule.Rule {
		return rule.Required{Attribute: attribute}
	})
}

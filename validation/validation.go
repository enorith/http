package validation

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/enorith/http/contracts"
	"github.com/enorith/http/validation/rule"
	"github.com/enorith/language"
)

type WithValidation interface {
	Rules() map[string][]interface{}
}

var DefaultValidator *Validator

type ValidateError map[string][]string

func (v ValidateError) StatusCode() int {
	return 422
}

func (v ValidateError) Error() string {
	var first string

	for _, s := range v {
		if len(s) > 0 {
			first = s[0]
		}
		break
	}

	return first
}

func Register(name string, register RuleRegister) {
	DefaultValidator.Register(name, register)
}

type RuleRegister func(attribute string, r contracts.InputSource, args ...string) (rule.Rule, error)

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
			inputRule, e := r(attribute, req, args...)
			if e != nil {
				errors = append(errors, e.Error())
				break
			}
			message, success, skip := v.passRule(inputRule, input, attribute, ss[0])
			if !success {
				errors = append(errors, message)
			}

			if skip {
				break
			}
		} else {
			return []string{fmt.Sprintf("unregisterd validation rule [%s]", ss[0])}
		}
	}
	return
}

func (v *Validator) PassesRules(req contracts.InputSource, attribute string, rules []interface{}) (errors []string) {
	input := req.GetValue(attribute)
	for i, rl := range rules {
		if s, ok := rl.(string); ok {
			ss := strings.Split(s, ":")
			var args []string
			if len(ss) > 1 {
				args = strings.Split(ss[1], ",")
			}
			r, exist := v.GetRule(ss[0])
			if exist {
				inputRule, e := r(attribute, req, args...)
				if e != nil {
					errors = append(errors, e.Error())
					break
				}
				message, success, skip := v.passRule(inputRule, input, attribute, ss[0])
				if !success {
					errors = append(errors, message)
				}

				if skip {
					break
				}
			}
		} else if rr, ok := rl.(rule.Rule); ok {
			message, success, skip := v.passRule(rr, input, attribute, fmt.Sprintf("rule %d", i))
			if skip {
				break
			}

			if !success {
				errors = append(errors, message)
			}
		}
	}

	return
}

func (v *Validator) passRule(r rule.Rule, input contracts.InputValue, attribute, name string) (string, bool, bool) {
	success, skip := r.Passes(input)

	var message string
	if !success {

		if msg, ok := r.(rule.Messager); ok {
			message = msg.Message()
		}
		if len(message) < 1 {
			var err error
			attr, _ := language.T("validation", "attributes."+attribute)
			if attr == "" {
				attr = attribute
			}

			message, err = language.T("validation", name, map[string]string{
				"attribute": attr,
			})
			if err != nil {
				message = fmt.Sprintf("validation attribute [%s] error, %s", attribute, name)
			}
		}
	}

	return message, success, skip
}

func init() {
	DefaultValidator = &Validator{registers: map[string]RuleRegister{}, mu: sync.RWMutex{}}
	Register("required", func(attribute string, r contracts.InputSource, args ...string) (rule.Rule, error) {
		return rule.Required{}, nil
	})

	Register("file", func(attribute string, r contracts.InputSource, args ...string) (rule.Rule, error) {
		return rule.FileInput{Attribute: attribute, Source: r}, nil
	})

	Register("nullable", func(attribute string, r contracts.InputSource, args ...string) (rule.Rule, error) {
		return rule.NullableInput{}, nil
	})

	Register("skipnull", func(attribute string, r contracts.InputSource, args ...string) (rule.Rule, error) {
		return rule.NullableInput{}, nil
	})

	Register("numeric", func(attribute string, r contracts.InputSource, args ...string) (rule.Rule, error) {
		if len(args) == 0 {
			return nil, errors.New("numeric rule require a numeric type, usage: validate:\"numeric:integer\"")
		}

		return rule.Numeric(rule.NumberType(args[0])), nil
	})

	Register("datetime", func(attribute string, r contracts.InputSource, args ...string) (rule.Rule, error) {
		return rule.Datetime(args...), nil
	})

	Register("in", func(attribute string, r contracts.InputSource, args ...string) (rule.Rule, error) {
		return rule.In(args...), nil
	})

	Register("required_if", func(attribute string, r contracts.InputSource, args ...string) (rule.Rule, error) {
		if len(args) == 0 {
			return nil, errors.New("required_if rule require a condition, usage: validate:\"required_if:field,value\"")
		}
		fv := r.Get(args[0])

		return rule.RequiredIf(func() bool {
			if len(args) > 1 {
				return bytes.Equal(fv, []byte(args[1]))
			}

			return len(fv) > 0
		}), nil
	})

	Register("required_with", func(attribute string, r contracts.InputSource, args ...string) (rule.Rule, error) {
		if len(args) == 0 {
			return nil, errors.New("required_with rule require a condition, usage: validate:\"required_with:field1,field2\"")
		}

		return rule.RequiredIf(func() bool {
			for _, v := range args {
				if len(r.Get(v)) > 0 {
					return true
				}
			}

			return false
		}), nil
	})
}

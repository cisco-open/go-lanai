package web

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var (
	bindingValidator = newValidator(binding.Validator)
)

// Validator returns the global validator for binding.
// Callers can register custom validators
func Validator() *Validate {
	return bindingValidator
}


func newValidator(ginValidator binding.StructValidator) *Validate {
	validate := ginValidator.Engine().(*validator.Validate)
	return &Validate{
		Validate: validate,
	}
}

// Validate is a thin wrapper around validator/v10, which prevent modifying TagName
type Validate struct {
	*validator.Validate
}

// WithTagName create a shallow copy of internal validator.Validate with different tag name
func (v *Validate) WithTagName(name string) *Validate {
	cp := Validate{
		Validate: v.Validate,
	}
	cp.Validate.SetTagName(name)
	return &cp
}

func (v *Validate) SetTagName(name string) {
	panic(fmt.Errorf("illegal attempt to modify tag of validator. Please use WithTagName(string)"))
}

// SetTranslations registers default translations using given regFn
func (v *Validate) SetTranslations(trans ut.Translator, regFn func(*validator.Validate, ut.Translator) error) error {
	return regFn(v.Validate, trans)
}


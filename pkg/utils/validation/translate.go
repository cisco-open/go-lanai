package validation

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

const (
	DefaultLocale = "en"
)

var (
	universalTranslator = newUniversalTranslator()
)

// DefaultTranslator returns the default ut.Translator of the package
func DefaultTranslator() ut.Translator {
	trans, _ := universalTranslator.GetTranslator(DefaultLocale)
	return trans
}

// UniversalTranslator returns the globally configured ut.UniversalTranslatorTranslator
// callers can register more locales
func UniversalTranslator() *ut.UniversalTranslator {
	return universalTranslator
}

// SimpleTranslationRegFunc returns a translation registration function for simple validation translate template
// the return ed function could used to register custom translation override
func SimpleTranslationRegFunc(tag, template string) func(*validator.Validate, ut.Translator) error {
	return func(validate *validator.Validate, trans ut.Translator) error {
		return validate.RegisterTranslation(tag, trans, func(ut ut.Translator) error {
			return ut.Add(tag, template, true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T(tag, fe.Field(), fe.Param())
			return t
		})
	}
}

func newUniversalTranslator() *ut.UniversalTranslator {
	english := en.New()
	return ut.New(english)
}

package controller

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/validation"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/go-playground/validator/v10/non-standard/validators"
	"go.uber.org/fx"
)

func Use() {
	bootstrap.AddOptions(
		fx.Invoke(register),
	)
}

func register(lc fx.Lifecycle, r *web.Registrar) {
	// validation, note, related validation translations are registered in errorhandling package
	_ = web.Validator().RegisterValidation("notblank", validators.NotBlank)
	_ = web.Validator().RegisterValidation("enumof", validation.CaseInsensitiveOneOf())
	_ = web.Validator().RegisterValidationCtx("date", validation.Regex("^\\d{4}-\\d{2}-\\d{2}$"))
	_ = web.Validator().RegisterValidationCtx("date-time", validation.Regex("^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}(?:\\.\\d+)?(?:Z|[\\+-]\\d{2}:\\d{2})?$"))
	_ = web.Validator().RegisterValidationCtx("uuid", validation.Regex("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$"))
	_ = web.Validator().RegisterValidationCtx("regexCD184", validation.Regex("^[a-zA-Z0-8-_=]{1,256}$"))
	_ = web.Validator().RegisterValidationCtx("regexEB33C", validation.Regex("^[a-zA-Z0-9-_=]{1,256}$"))
	_ = web.Validator().RegisterValidationCtx("regexA79C5", validation.Regex("^[a-zA-Z0-5-_=]{1,256}$"))
	_ = web.Validator().RegisterValidationCtx("regexA397E", validation.Regex("^[a-zA-Z0-7-_=]{1,256}$"))
}

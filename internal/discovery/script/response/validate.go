package response

import (
	"github.com/cherts/pgscv/internal/validators"
	"github.com/go-playground/validator/v10"
)

// Validate performs validation on the ScriptResponse struct using custom and built-in validators.
// It checks field requirements, formats, and custom validation rules.
//
// Returns:
//   - error: if validation fails, containing details about validation errors
func (c *ScriptResponse) Validate() error {
	validate := validator.New()

	registerCustomValidators(validate)

	err := validate.Struct(c)
	if err != nil {
		return err
	}

	return nil
}

func registerCustomValidators(v *validator.Validate) {
	_ = v.RegisterValidation(validators.EnvNameValidator, validators.EnvNameValidatorFunc)
}

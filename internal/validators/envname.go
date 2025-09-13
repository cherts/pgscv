package validators

import (
	"github.com/go-playground/validator/v10"
)

const EnvNameValidator = "env_name"

func EnvNameValidatorFunc(fl validator.FieldLevel) bool {
	envVarName := fl.Field().String()

	if envVarName == "" {
		return false
	}

	if len(envVarName) == 0 {
		return false
	}

	firstChar := envVarName[0]
	if !(firstChar >= 'A' && firstChar <= 'Z') &&
		!(firstChar >= 'a' && firstChar <= 'z') &&
		firstChar != '_' {
		return false
	}

	for i := 1; i < len(envVarName); i++ {
		char := envVarName[i]
		if !(char >= 'A' && char <= 'Z') &&
			!(char >= 'a' && char <= 'z') &&
			!(char >= '0' && char <= '9') &&
			char != '_' {
			return false
		}
	}

	return true
}

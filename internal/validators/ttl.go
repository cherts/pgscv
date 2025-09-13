package validators

import (
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
)

const TTLValidator = "ttl"

func TTLValidatorFunc(fl validator.FieldLevel) bool {
	ttlStr := fl.Field().String()

	if ttlStr == "" {
		return false
	}

	duration, err := time.ParseDuration(ttlStr)
	if err != nil {
		if seconds, err := strconv.Atoi(ttlStr); err == nil {
			return seconds > 0
		}
		return false
	}

	return duration > 0
}

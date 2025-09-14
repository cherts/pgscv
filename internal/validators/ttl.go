package validators

import (
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
)

// TTLValidatorFunc validates that a string represents a valid positive time duration.
// The function accepts both Go duration strings (e.g., "30s", "1m") and integer seconds.
//
// Parameters:
//   - fl: FieldLevel containing the field to validate
//
// Returns:
//   - bool: true if the field represents a positive time duration, false otherwise
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

// Package validators provides custom validation functions for use with the go-playground/validator package.
// It includes validators for common patterns used in the pgSCV monitoring system, such as:
//
// - Environment variable name validation
// - Regular file path validation
// - Time-to-live (TTL) duration validation
//
// These validators are designed to be registered with validator.v10 and used in struct field tags
// to enforce specific format requirements and constraints.
package validators

const (
	// EnvNameValidator is the tag name used for environment variable name validation
	EnvNameValidator = "env_name"
	// RegularFileValidator is the tag name used for regular file validation
	RegularFileValidator = "regular_file"
	// TTLValidator is the tag name used for Time-To-Live (TTL) validation
	TTLValidator = "ttl"
)

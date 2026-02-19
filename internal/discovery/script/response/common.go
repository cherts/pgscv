// Package response provides functionality for unmarshal script output into structured Go data types and validate this.
// It is designed to parse formatted text output from discovery scripts
// and convert it into slices of structs using reflection and struct tags for mapping.
// The parser in plain format handles hyphens (-) as empty values and missing fields gracefully.
package response

import "errors"

// Common error variables used throughout the package
var (
	errOutputParameterMustBePointerToSlice = errors.New("output parameter must be a pointer to a slice")
	errSliceElementMustBeStruct            = errors.New("slice element must be a struct")
	errNoDataProvided                      = errors.New("no data provided")
)

// ScriptResponse represents the structured response from a discovery script.
// It contains connection information and credentials for PostgreSQL database monitoring.
type ScriptResponse struct {
	// ServiceID is the unique identifier for the service being monitored
	ServiceID string `pgscv:"service-id" json:"service_id" validate:"required"`
	// DSN is the PostgreSQL connection string in libpq format
	//
	// @todo pg_dsn validator, not important: see internal/discovery/service/script.go
	//
	// func fillSvcResponse(svc *response.ScriptResponse) error {
	//
	//     dbConfig, err = pgx.ParseConfig(svc.DSN)
	//
	//     if err != nil {
	//		    return err <-----
	//     }
	DSN string `pgscv:"dsn" json:"dsn"`
	// Database is the database name
	Database string `pgscv:"database" json:"database"`
	// Host is the database server hostname or IP address
	Host string `pgscv:"host" json:"host"`
	// Port is the database server port number
	Port uint16 `pgscv:"port" json:"port" validate:"omitempty,port"`
	// User is the database username for connection
	User string `pgscv:"user" json:"user"`
	// UserFromEnv specifies an environment variable containing the database username
	UserFromEnv string `pgscv:"user-from-env" json:"user_from_env" validate:"omitempty,env_name"`
	// Password is the database password for connection
	Password string `pgscv:"password" json:"password"` // #nosec G117
	// PasswordFromEnvVar specifies an environment variable containing the database password
	PasswordFromEnvVar string `pgscv:"password-from-env" json:"password_from_env" validate:"omitempty,env_name"`
}

// AllFieldsEmpty checks if all fields in the ScriptResponse are empty/zero values.
// Returns true if all fields are empty, false otherwise.
func (c *ScriptResponse) AllFieldsEmpty() bool {
	return c.ServiceID == "" && c.DSN == "" && c.Host == "" && c.Port == 0 && c.User == "" &&
		c.UserFromEnv == "" && c.Password == "" && c.PasswordFromEnvVar == ""
}

package response

import "errors"

var (
	errOutputParameterMustBePointerToSlice = errors.New("output parameter must be a pointer to a slice")
	errSliceElementMustBeStruct            = errors.New("slice element must be a struct")
	errNoDataProvided                      = errors.New("no data provided")
)

type ScriptResponse struct {
	ServiceID string `pgscv:"service-id" json:"service_id" validate:"required"`
	// @todo pg_dsn validator, not important: see internal/discovery/service/script.go
	// func fillSvcResponse(svc *response.ScriptResponse) error {
	// ...
	// dbConfig, err = pgx.ParseConfig(svc.DSN)
	// if err != nil {
	//		return err <-----
	// }
	DSN                string `pgscv:"dsn" json:"dsn"`
	Host               string `pgscv:"host" json:"host"`
	Port               uint16 `pgscv:"port" json:"port" validate:"omitempty,port"`
	User               string `pgscv:"user" json:"user"`
	UserFromEnv        string `pgscv:"user-from-env" json:"user_from_env" validate:"omitempty,env_name"`
	Password           string `pgscv:"password" json:"password"`
	PasswordFromEnvVar string `pgscv:"password-from-env" json:"password_from_env" validate:"omitempty,env_name"`
}

func (c *ScriptResponse) AllFieldsEmpty() bool {
	return c.ServiceID == "" && c.DSN == "" && c.Host == "" && c.Port == 0 && c.User == "" &&
		c.UserFromEnv == "" && c.Password == "" && c.PasswordFromEnvVar == ""
}

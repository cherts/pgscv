// Package response provides functionality for unmarshal script output into structured Go data types and validate this.
// It is designed to parse formatted text output from discovery scripts
// and convert it into slices of structs using reflection and struct tags for mapping.
// The parser in plain format handles hyphens (-) as empty values and missing fields gracefully.
package response

import (
	"fmt"
	"github.com/cherts/pgscv/internal/log"
	"reflect"
	"strconv"
	"strings"
)

const scriptResponseReflectKey = "pgscv"

// UnmarshalScriptResponse unmarshal plain script response using reflect.
// @todo: format json, specifying delimiter for plain format
// Example input:
/*
# service-id dsn password-from-env password
ubuntu_24_main_16 postgres://exporter@ubuntu-host:7432/postgres PGPASSWORD -
ubuntu_24_main_16-slave postgres://exporter@ubuntu-slave-host:7432/postgres - my_secret_password123
*/
func UnmarshalScriptResponse(data string, out any) error {
	var (
		fieldPositions = map[string]int{}
	)

	v := reflect.ValueOf(out)
	if v.Kind() != reflect.Ptr {
		return errOutputParameterMustBePointerToSlice
	}

	sliceVal := v.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return errOutputParameterMustBePointerToSlice
	}

	elemType := sliceVal.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return errSliceElementMustBeStruct
	}

	lines := strings.Split(strings.TrimSpace(data), "\n")
	if len(lines) == 0 {
		return errNoDataProvided
	}

	results := reflect.MakeSlice(sliceVal.Type(), 0, len(lines)-1)
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			fieldPositions = make(map[string]int)

			headerLine := strings.TrimSpace(line)
			headerFields := strings.Fields(headerLine[1:])

			for i, field := range headerFields {
				fieldPositions[field] = i
			}

			continue
		}

		if len(fieldPositions) == 0 {
			continue
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < len(fieldPositions) {
			return fmt.Errorf("line has fewer fields than header: %s", line)
		}

		elem := reflect.New(elemType).Elem()

		for i := 0; i < elemType.NumField(); i++ {
			field := elemType.Field(i)

			tag := field.Tag.Get(scriptResponseReflectKey)
			if tag == "" {
				continue
			}

			pos, exists := fieldPositions[tag]
			if !exists {
				continue
			}

			if pos >= len(fields) {
				continue
			}

			fieldValue := fields[pos]
			if fieldValue == "-" {
				fieldValue = ""
			}

			fieldVal := elem.Field(i)
			if !fieldVal.CanSet() {
				continue
			}

			switch fieldVal.Kind() {
			case reflect.String:
				fieldVal.SetString(fieldValue)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				intVal, err := strconv.ParseInt(fieldValue, 10, 64)
				if err != nil {
					log.Errorf("failed to parse int value: %s", fieldValue)

					intVal = 0
				}

				fieldVal.SetInt(intVal)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				uintVal, err := strconv.ParseUint(fieldValue, 10, 64)
				if err != nil {
					log.Errorf("failed to parse uint value: %s", fieldValue)

					uintVal = 0
				}

				fieldVal.SetUint(uintVal)
			case reflect.Bool:
				boolVal, err := strconv.ParseBool(fieldValue)
				if err != nil {
					log.Errorf("failed to parse bool value: %s", fieldValue)

					boolVal = false
				}

				fieldVal.SetBool(boolVal)
			case reflect.Float32, reflect.Float64:
				floatVal, err := strconv.ParseFloat(fieldValue, 64)
				if err != nil {
					log.Errorf("failed to parse float value: %s", fieldValue)

					floatVal = 0
				}

				fieldVal.SetFloat(floatVal)
			default:
				return fmt.Errorf("unsupported field type: %s", fieldVal.Kind())
			}
		}

		results = reflect.Append(results, elem)
	}

	sliceVal.Set(results)

	return nil
}

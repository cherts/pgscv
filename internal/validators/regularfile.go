package validators

import (
	"os"
	"path/filepath"

	"github.com/cherts/pgscv/internal/log"
	"github.com/go-playground/validator/v10"
)

// RegularFileValidatorFunc validates that a string path points to a regular file (not a directory, symlink, etc).
//
// Parameters:
//   - fl: FieldLevel containing the field to validate
//
// Returns:
//   - bool: true if the path points to a regular file, false otherwise
func RegularFileValidatorFunc(fl validator.FieldLevel) bool {
	f := fl.Field().String()
	if f == "" {
		return false
	}

	fileInfo, err := os.Lstat(filepath.Clean(f))
	if err != nil {
		log.Errorf("failed to lstat file: %s %v", f, err)
		return false
	}

	return fileInfo.Mode().IsRegular()
}

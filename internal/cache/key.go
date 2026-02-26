package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/cherts/pgscv/internal/log"
	"strings"
)

func hash(args ...any) string {
	var s strings.Builder
	for _, v := range args {
		s.WriteString(fmt.Sprintf("%v", v))
	}

	hash := sha256.New()
	_, err := hash.Write([]byte(s.String()))
	if err != nil {
		log.Error(fmt.Sprintf("Error calculating hash: %v", err))
		return ""
	}
	hashKey := hex.EncodeToString(hash.Sum(nil))
	return hashKey
}

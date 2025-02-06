package yandex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cherts/pgscv/discovery/log"
	"github.com/golang-jwt/jwt/v4"
)

// authorizedKey describe authorized key json structure
type authorizedKey struct {
	ID               string `json:"id"`
	ServiceAccountID string `json:"service_account_id"`
	CreatedAt        string `json:"created_at"`
	KeyAlgorithm     string `json:"key_algorithm"`
	PublicKey        string `json:"public_key"`
	PrivateKey       string `json:"private_key"`
}

type tokenIAM struct {
	sync.RWMutex
	IAMToken  string `json:"iamToken"`
	ExpiresAt string `json:"expiresAt"`
	key       authorizedKey
}

func newIAMToken(jsonFilePath string) (*tokenIAM, error) {
	token := &tokenIAM{}
	err := token.loadAuthorizedKey(jsonFilePath)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (token *tokenIAM) GetToken() (*string, error) {
	log.Debug("[SD] Getting IAM token...")
	if token.IsExpired() {
		err := token.Renew()
		if err != nil {
			return nil, err
		}
	}
	token.RLock()
	defer token.RUnlock()
	return &token.IAMToken, nil
}

func (token *tokenIAM) IsExpired() bool {
	token.RLock()
	defer token.RUnlock()
	if token.ExpiresAt == "" {
		return true
	}

	t, err := time.Parse(time.RFC3339, token.ExpiresAt)
	if err != nil {
		return true
	}
	return t.Before(time.Now().Add(time.Duration(-30) * time.Minute))
}

func (token *tokenIAM) Renew() error {
	log.Debug("[SD] Renewing IAM token...")
	token.Lock()
	defer token.Unlock()
	jwtToken, err := token.getJWTToken()
	if err != nil {
		return err
	}

	// See https://yandex.cloud/ru/docs/iam/api-ref/IamToken/create
	resp, err := http.Post(
		"https://iam.api.cloud.yandex.net/iam/v1/tokens",
		"application/json",
		strings.NewReader(fmt.Sprintf(`{"jwt":"%s"}`, *jwtToken)),
	)
	if err != nil {
		log.Errorf("[SD] Failed to renew IAM token, error: %s", err.Error())
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Errorf("[SD] Failed to renew IAM token, unexpected status code: %d", resp.StatusCode)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(token); err != nil {
		log.Errorf("[SD] Failed to renew IAM token, decode error: %s", err.Error())
		return err
	}
	return nil
}

func (token *tokenIAM) getJWTToken() (*string, error) {
	log.Debug("[SD] Getting JWT token...")
	rsaPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(token.key.PrivateKey))
	if err != nil {
		return nil, err
	}
	claims := jwt.RegisteredClaims{
		Issuer:    token.key.ServiceAccountID,
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		Audience:  []string{"https://iam.api.cloud.yandex.net/iam/v1/tokens"},
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodPS256, claims)
	jwtToken.Header["kid"] = token.key.ID
	tokenString, err := jwtToken.SignedString(rsaPrivateKey)
	if err != nil {
		log.Errorf("[SD] Failed to get JWT token, error: %s", err.Error())
		return nil, err
	}
	return &tokenString, nil
}

func (token *tokenIAM) loadAuthorizedKey(filePath string) error {
	log.Debugf("[SD] Loading authorized key from path '%s'", filePath)
	data, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		log.Errorf("[SD] Failed to load authorized key, error: %s", err.Error())
		return err
	}
	err = json.Unmarshal(data, &token.key)
	if err != nil {
		log.Errorf("[SD] Failed to parse authorized key, JSON parse error: %s", err.Error())
		return err
	}
	return nil
}

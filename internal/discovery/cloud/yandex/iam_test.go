package yandex

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestIsExpired(t *testing.T) {
	type args struct {
		tokenString string
		expiresAt   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "token expired",
			args: args{
				tokenString: generateToken(time.Now().Add(-1 * time.Hour)),
				expiresAt:   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			},
			want: true,
		},
		{
			name: "token not expired",
			args: args{
				tokenString: generateToken(time.Now().Add(11 * time.Hour)),
				expiresAt:   time.Now().Add(11 * time.Hour).Format(time.RFC3339),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &tokenIAM{
				IAMToken:  tt.args.tokenString,
				ExpiresAt: tt.args.expiresAt,
			}

			got := token.IsExpired()
			assert.Equalf(t, tt.want, got, "IsExpired(%v)", tt.args.tokenString)
		})
	}
}

func generateToken(expTime time.Time) string {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expTime),
		Issuer:    "your-issuer",
		Subject:   "your-subject",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte("your-secret"))
	if err != nil {
		return ""
	}

	return tokenString
}

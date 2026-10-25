package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"app/internal/infrastructure/http/response"
)

const (
	minimumJWTSecretLength = 32
	subjectContextKey      = "auth.jwt.subject"
)

var (
	ErrJWTSecretMissing  = errors.New("JWT secret is required")
	ErrJWTSecretTooShort = errors.New("JWT secret must contain at least 32 bytes")
)

// NewJWT creates middleware that authenticates Bearer tokens signed with HS256.
func NewJWT(secret string) (gin.HandlerFunc, error) {
	if secret == "" {
		return nil, ErrJWTSecretMissing
	}
	if len([]byte(secret)) < minimumJWTSecretLength {
		return nil, ErrJWTSecretTooShort
	}

	verificationKey := []byte(secret)

	return func(ctx *gin.Context) {
		tokenString, ok := bearerToken(ctx.Request.Header.Values("Authorization"))
		if !ok {
			unauthorized(ctx)
			return
		}

		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(
			tokenString,
			claims,
			func(token *jwt.Token) (any, error) {
				return verificationKey, nil
			},
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
			jwt.WithExpirationRequired(),
		)
		if err != nil || !token.Valid {
			unauthorized(ctx)
			return
		}

		if claims.Subject != "" {
			ctx.Set(subjectContextKey, claims.Subject)
		}
		ctx.Next()
	}, nil
}

// Subject returns the authenticated token subject when the claim was supplied.
func Subject(ctx *gin.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	subject, ok := ctx.Get(subjectContextKey)
	if !ok {
		return "", false
	}

	value, ok := subject.(string)
	return value, ok && value != ""
}

func bearerToken(headers []string) (string, bool) {
	if len(headers) != 1 {
		return "", false
	}

	parts := strings.Fields(headers[0])
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

func unauthorized(ctx *gin.Context) {
	ctx.Header("WWW-Authenticate", "Bearer")
	response.Error(ctx, http.StatusUnauthorized, "UNAUTHORIZED", "token de autenticação ausente ou inválido", nil)
	ctx.Abort()
}

package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"app/internal/infrastructure/http/middleware"
)

const (
	testSecret      = "test-secret-with-at-least-32-bytes"
	differentSecret = "different-secret-with-at-least-32-bytes"
)

func TestNewJWTValidatesSecret(t *testing.T) {
	_, err := middleware.NewJWT("")
	require.ErrorIs(t, err, middleware.ErrJWTSecretMissing)

	_, err = middleware.NewJWT("short")
	require.ErrorIs(t, err, middleware.ErrJWTSecretTooShort)

	result, err := middleware.NewJWT(testSecret)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestJWTAuthenticatesAndPropagatesSubject(t *testing.T) {
	token := signedToken(t, jwt.SigningMethodHS256, []byte(testSecret), jwt.RegisteredClaims{
		Subject:   "user-123",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})

	response := performProtectedRequest(t, []string{"Bearer " + token})

	require.Equal(t, http.StatusOK, response.Code)
	var payload map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.Equal(t, "user-123", payload["subject"])
}

func TestJWTAllowsTokenWithoutSubject(t *testing.T) {
	token := signedToken(t, jwt.SigningMethodHS256, []byte(testSecret), jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})

	response := performProtectedRequest(t, []string{"Bearer " + token})

	require.Equal(t, http.StatusOK, response.Code)
	var payload map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.Equal(t, "", payload["subject"])
}

func TestJWTRejectsInvalidAuthentication(t *testing.T) {
	validClaims := jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}
	expiredClaims := jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour))}
	futureClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		NotBefore: jwt.NewNumericDate(time.Now().Add(time.Minute)),
	}

	hmac384 := signedToken(t, jwt.SigningMethodHS384, []byte(testSecret), validClaims)
	unsigned := signedToken(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, validClaims)

	tests := []struct {
		name    string
		headers []string
	}{
		{name: "missing header"},
		{name: "duplicate header", headers: []string{"Bearer token-one", "Bearer token-two"}},
		{name: "wrong scheme", headers: []string{"Basic credentials"}},
		{name: "missing token", headers: []string{"Bearer"}},
		{name: "extra header field", headers: []string{"Bearer token extra"}},
		{name: "malformed token", headers: []string{"Bearer malformed"}},
		{name: "wrong signature", headers: []string{"Bearer " + signedToken(t, jwt.SigningMethodHS256, []byte(differentSecret), validClaims)}},
		{name: "expired token", headers: []string{"Bearer " + signedToken(t, jwt.SigningMethodHS256, []byte(testSecret), expiredClaims)}},
		{name: "missing expiration", headers: []string{"Bearer " + signedToken(t, jwt.SigningMethodHS256, []byte(testSecret), jwt.RegisteredClaims{})}},
		{name: "not active yet", headers: []string{"Bearer " + signedToken(t, jwt.SigningMethodHS256, []byte(testSecret), futureClaims)}},
		{name: "HS384 algorithm", headers: []string{"Bearer " + hmac384}},
		{name: "none algorithm", headers: []string{"Bearer " + unsigned}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := performProtectedRequest(t, test.headers)

			require.Equal(t, http.StatusUnauthorized, response.Code)
			require.Equal(t, "Bearer", response.Header().Get("WWW-Authenticate"))

			var payload map[string]any
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
			require.Equal(t, false, payload["success"])
			require.Equal(t, float64(http.StatusUnauthorized), payload["status"])
			errorPayload := payload["error"].(map[string]any)
			require.Equal(t, "UNAUTHORIZED", errorPayload["code"])
			require.Equal(t, "token de autenticação ausente ou inválido", errorPayload["message"])
			require.NotContains(t, errorPayload, "details")
		})
	}
}

func performProtectedRequest(t *testing.T, headers []string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	authMiddleware, err := middleware.NewJWT(testSecret)
	require.NoError(t, err)

	engine := gin.New()
	engine.GET("/protected", authMiddleware, func(ctx *gin.Context) {
		subject, _ := middleware.Subject(ctx)
		ctx.JSON(http.StatusOK, gin.H{"subject": subject})
	})

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	for _, header := range headers {
		request.Header.Add("Authorization", header)
	}
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	return response
}

func signedToken(t *testing.T, method jwt.SigningMethod, key any, claims jwt.RegisteredClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(method, claims)
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

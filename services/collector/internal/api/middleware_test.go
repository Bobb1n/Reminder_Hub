package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestUserIDAuth_ValidUUID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	validUUID := uuid.New().String()
	req.Header.Set("X-User-ID", validUUID)

	handler := UserIDAuth(func(c echo.Context) error {
		userID := c.Get(ContextKeyUserID)
		assert.Equal(t, validUUID, userID)
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUserIDAuth_MissingHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := UserIDAuth(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)

	assert.Error(t, err)
	httpErr := err.(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}

func TestUserIDAuth_InvalidUUID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	req.Header.Set("X-User-ID", "invalid-uuid")

	handler := UserIDAuth(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)

	assert.Error(t, err)
	httpErr := err.(*echo.HTTPError)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestIsValidUUID_Valid(t *testing.T) {
	validUUID := uuid.New().String()
	assert.True(t, isValidUUID(validUUID))
}

func TestIsValidUUID_InvalidLength(t *testing.T) {
	assert.False(t, isValidUUID("short"))
	assert.False(t, isValidUUID("too-long-uuid-that-exceeds-36-characters"))
}

func TestIsValidUUID_InvalidFormat(t *testing.T) {
	assert.False(t, isValidUUID("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")) // Неправильные символы
	assert.False(t, isValidUUID("12345678-1234-1234-1234-123456789012")) // Правильный формат
	assert.True(t, isValidUUID("12345678-1234-1234-1234-123456789012"))
}

func TestIsValidUUID_WrongSeparators(t *testing.T) {
	assert.False(t, isValidUUID("12345678_1234-1234-1234-123456789012")) // Неправильный разделитель
}

func TestInternalAuth_ValidToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	internalToken := "test-token"
	req.Header.Set("Authorization", "Bearer test-token")

	handler := InternalAuth(internalToken)(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestInternalAuth_MissingHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := InternalAuth("test-token")(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)

	assert.NoError(t, err)
	// Проверяем, что вернулся JSON с ошибкой
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestInternalAuth_InvalidFormat(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	req.Header.Set("Authorization", "InvalidFormat")

	handler := InternalAuth("test-token")(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestInternalAuth_WrongToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	req.Header.Set("Authorization", "Bearer wrong-token")

	handler := InternalAuth("correct-token")(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

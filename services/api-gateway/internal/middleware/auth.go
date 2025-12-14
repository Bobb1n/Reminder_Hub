package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(authServiceURL string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Path() == "/health" || strings.HasPrefix(c.Path(), "/auth/") {
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Authorization header required")
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization format")
			}

			token := parts[1]
			userID, err := validateToken(authServiceURL, token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}
			c.Set("user_id", userID)
			return next(c)
		}
	}
}

type validateTokenResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func validateToken(authServiceURL, token string) (string, error) {
	validateURL := strings.TrimSuffix(authServiceURL, "/") + "/auth/validate"

	resp, err := sendValidateRequest(validateURL, token)
	if err != nil {
		return "", fmt.Errorf("validate request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", handleErrorResponse(resp)
	}

	return parseValidateResponse(resp.Body)
}

// sendValidateRequest отправляет запрос на валидацию токена
func sendValidateRequest(validateURL, token string) (*http.Response, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, validateURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return resp, nil
}

// handleErrorResponse обрабатывает ошибочный ответ от сервера
func handleErrorResponse(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("auth service returned status %d: %s", resp.StatusCode, string(body))
}

// parseValidateResponse парсит и валидирует успешный ответ
func parseValidateResponse(body io.Reader) (string, error) {
	var tokenResp validateTokenResponse
	if err := json.NewDecoder(body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if !tokenResp.Valid {
		return "", fmt.Errorf("token is not valid")
	}

	if tokenResp.UserID == "" {
		return "", fmt.Errorf("user_id is empty in response")
	}

	return tokenResp.UserID, nil
}

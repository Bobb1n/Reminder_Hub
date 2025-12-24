package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"reminder-hub/pkg/logger"

	"github.com/labstack/echo/v4"
)

type ServiceProxy struct {
	targetURL     *url.URL
	internalToken string
	logger        *logger.CurrentLogger
	serviceType   string
}

func NewServiceProxy(targetURL, internalToken string, log *logger.CurrentLogger) (*ServiceProxy, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	serviceType := "core"
	if strings.Contains(targetURL, "collector") {
		serviceType = "collector"
	}

	return &ServiceProxy{
		targetURL:     parsedURL,
		internalToken: internalToken,
		logger:        log,
		serviceType:   serviceType,
	}, nil
}

func (p *ServiceProxy) Proxy(c echo.Context) error {
	proxy := httputil.NewSingleHostReverseProxy(p.targetURL)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Переписываем путь для integrations
		// Используем реальный путь запроса, а не шаблон маршрута
		requestPath := c.Request().URL.Path
		ctx := c.Request().Context()

		if strings.HasPrefix(requestPath, "/api/v1/integrations/email") {
			if c.Request().Method == http.MethodGet {
				// Для GET запросов подставляем user_id в путь
				if userID, ok := c.Get("user_id").(string); ok && userID != "" {
					newPath := "/api/integrations/" + userID
					req.URL.Path = newPath
					p.logger.Debug(ctx, "Rewritten path for GET", "from", requestPath, "to", newPath)
				} else {
					// Если user_id не найден, логируем ошибку
					p.logger.Warn(ctx, "user_id not found in context for GET /api/v1/integrations/email")
				}
			} else {
				// Для POST, PUT, DELETE запросов убираем /v1/integrations/email
				req.URL.Path = "/api/integrations"
				p.logger.Debug(ctx, "Rewritten path", "method", c.Request().Method, "from", requestPath, "to", "/api/integrations")
			}
		}

		// Переписываем путь для reminders -> tasks (только для collector-service)
		if p.serviceType == "collector" && strings.HasPrefix(requestPath, "/api/v1/reminders") {
			// Заменяем /api/v1/reminders на /api/v1/tasks
			// Убираем /reminders из пути
			req.URL.Path = "/api/v1/tasks"
			p.logger.Debug(ctx, "Rewritten path for reminders", "from", requestPath, "to", req.URL.Path)
		}

		// Копируем body из модифицированного запроса (после middleware)
		// Это важно, если middleware изменил body (например, AutoIMAPMiddleware)
		if c.Request().Body != nil && (c.Request().Method == http.MethodPost || c.Request().Method == http.MethodPut || c.Request().Method == http.MethodPatch) {
			bodyBytes, err := io.ReadAll(c.Request().Body)
			if err == nil && len(bodyBytes) > 0 {
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				req.ContentLength = int64(len(bodyBytes))
				req.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes)))
				// Восстанавливаем body в оригинальном запросе для возможного повторного использования
				c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				p.logger.Debug(ctx, "Copied body to proxy request", "length", len(bodyBytes))
			}
		}

		// Для core-service передаем внутренний токен в Authorization header
		// (core-service ожидает внутренний токен, а не токен пользователя)
		req.Header.Set("Authorization", "Bearer "+p.internalToken)

		// Также добавляем внутренний токен в отдельный header для совместимости
		req.Header.Set("X-Internal-Token", p.internalToken)

		if userID, ok := c.Get("user_id").(string); ok {
			req.Header.Set("X-User-ID", userID)
		}

		req.Header.Set("X-Forwarded-By", "api-gateway")
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		c.Logger().Errorf("Proxy error: %v", err)
		c.JSON(http.StatusBadGateway, map[string]string{
			"error": "Service unavailable",
		})
	}

	proxy.ServeHTTP(c.Response(), c.Request())
	return nil
}

func AuthProxy(targetURL string) (echo.HandlerFunc, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)

	return func(c echo.Context) error {
		proxy.ServeHTTP(c.Response(), c.Request())
		return nil
	}, nil
}

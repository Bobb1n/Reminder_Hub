package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"reminder-hub/pkg/logger"

	"github.com/labstack/echo/v4"
)

type IMAPSettings struct {
	Host string
	Port int
	SSL  bool
}

// AutoIMAPMiddleware автоматически определяет IMAP настройки по email
func AutoIMAPMiddleware(log *logger.CurrentLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			requestPath := c.Request().URL.Path
			method := c.Request().Method

			log.Debug(ctx, "AutoIMAPMiddleware: checking request", "method", method, "path", requestPath)

			// Применяем только к POST запросам на создание интеграции
			if method != http.MethodPost {
				log.Debug(ctx, "AutoIMAPMiddleware: skipping non-POST request")
				return next(c)
			}

			// Проверяем, что это запрос на создание интеграции
			// Используем реальный путь запроса, а не шаблон маршрута
			if !strings.Contains(requestPath, "/integrations/email") {
				log.Debug(ctx, "AutoIMAPMiddleware: skipping non-integrations path")
				return next(c)
			}

			// Читаем body
			bodyBytes, err := io.ReadAll(c.Request().Body)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Failed to read request body")
			}
			c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Парсим JSON
			var reqBody map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
				return next(c) // Если не JSON, пропускаем
			}

			// Проверяем, есть ли email_address
			email, ok := reqBody["email_address"].(string)
			if !ok || email == "" {
				return next(c) // Если нет email, пропускаем
			}

			log.Info(ctx, "AutoIMAPMiddleware: processing email", "email", email, "original_body_length", len(bodyBytes))

			// Если IMAP настройки не указаны - определяем автоматически
			imapHost, hasHost := reqBody["imap_host"].(string)
			if !hasHost || imapHost == "" {
				settings := getIMAPSettingsByEmail(email)

				reqBody["imap_host"] = settings.Host
				reqBody["imap_port"] = float64(settings.Port) // JSON числа всегда float64
				reqBody["use_ssl"] = settings.SSL
			} else {
				// Если host указан, но port нет - добавляем дефолтный порт
				if reqBody["imap_port"] == nil {
					reqBody["imap_port"] = float64(993)
				}
				// Если use_ssl не указан - добавляем дефолтное значение
				if reqBody["use_ssl"] == nil {
					reqBody["use_ssl"] = true
				}
			}

			// Сериализуем обратно в JSON
			newBodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process request")
			}

			// Заменяем body
			c.Request().Body = io.NopCloser(bytes.NewBuffer(newBodyBytes))
			c.Request().ContentLength = int64(len(newBodyBytes))

			// Обновляем Content-Length header
			c.Request().Header.Set("Content-Length", fmt.Sprintf("%d", len(newBodyBytes)))

			log.Info(ctx, "AutoIMAPMiddleware: modified body", "original_length", len(bodyBytes), "new_length", len(newBodyBytes), "content", string(newBodyBytes))
			return next(c)
		}
	}
}

// getIMAPSettingsByEmail определяет IMAP настройки по домену email
func getIMAPSettingsByEmail(email string) IMAPSettings {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return getDefaultIMAPSettings(email)
	}

	domain := strings.ToLower(strings.TrimSpace(parts[1]))

	// Switch по популярным провайдерам
	switch domain {
	case "gmail.com":
		return IMAPSettings{
			Host: "imap.gmail.com",
			Port: 993,
			SSL:  true,
		}
	case "outlook.com", "hotmail.com", "live.com", "msn.com":
		return IMAPSettings{
			Host: "outlook.office365.com",
			Port: 993,
			SSL:  true,
		}
	case "yandex.ru", "yandex.com":
		return IMAPSettings{
			Host: "imap.yandex.ru",
			Port: 993,
			SSL:  true,
		}
	case "mail.ru", "inbox.ru", "list.ru", "bk.ru":
		return IMAPSettings{
			Host: "imap.mail.ru",
			Port: 993,
			SSL:  true,
		}
	case "yahoo.com", "yahoo.co.uk", "yahoo.fr":
		return IMAPSettings{
			Host: "imap.mail.yahoo.com",
			Port: 993,
			SSL:  true,
		}
	case "protonmail.com", "proton.me":
		return IMAPSettings{
			Host: "127.0.0.1", // ProtonMail использует Bridge
			Port: 1143,
			SSL:  false,
		}
	default:
		// Для неизвестных провайдеров пробуем стандартные настройки
		return getDefaultIMAPSettings(domain)
	}
}

// getDefaultIMAPSettings возвращает дефолтные IMAP настройки
func getDefaultIMAPSettings(domain string) IMAPSettings {
	return IMAPSettings{
		Host: "imap." + domain,
		Port: 993,
		SSL:  true,
	}
}

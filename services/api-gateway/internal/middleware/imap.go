package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type IMAPSettings struct {
	Host string
	Port int
	SSL  bool
}

// AutoIMAPMiddleware автоматически определяет IMAP настройки по email
func AutoIMAPMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Применяем только к POST запросам на создание интеграции
			if c.Request().Method != http.MethodPost {
				return next(c)
			}

			// Проверяем, что это запрос на создание интеграции
			if !strings.Contains(c.Path(), "/integrations/email") {
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

			// Если IMAP настройки не указаны - определяем автоматически
			if reqBody["imap_host"] == nil || reqBody["imap_host"] == "" {
				settings := getIMAPSettingsByEmail(email)
				
				reqBody["imap_host"] = settings.Host
				reqBody["imap_port"] = settings.Port
				reqBody["use_ssl"] = settings.SSL
			}

			// Сериализуем обратно в JSON
			newBodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process request")
			}

			// Заменяем body
			c.Request().Body = io.NopCloser(bytes.NewBuffer(newBodyBytes))
			c.Request().ContentLength = int64(len(newBodyBytes))

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

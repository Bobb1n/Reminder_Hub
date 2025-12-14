package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/labstack/echo/v4"
)

type ServiceProxy struct {
	targetURL     *url.URL
	internalToken string
}

func NewServiceProxy(targetURL, internalToken string) (*ServiceProxy, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	return &ServiceProxy{
		targetURL:     parsedURL,
		internalToken: internalToken,
	}, nil
}

func (p *ServiceProxy) Proxy(c echo.Context) error {
	proxy := httputil.NewSingleHostReverseProxy(p.targetURL)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Передаем Authorization header от клиента
		if authHeader := c.Request().Header.Get("Authorization"); authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}

		// Добавляем внутренний токен
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

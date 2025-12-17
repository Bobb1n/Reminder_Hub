package configurations

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestConfigMiddlewares_Registers(t *testing.T) {
	e := echo.New()
	ConfigMiddlewares(e)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req) // just ensure no panic and pipeline exists
}

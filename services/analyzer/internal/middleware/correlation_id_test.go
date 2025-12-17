package echomiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCorrelationIdMiddleware_SetsHeaderAndContext(t *testing.T) {
	e := echo.New()
	h := CorrelationIdMiddleware(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	id := rec.Header().Get(echo.HeaderXCorrelationID)
	if id == "" {
		t.Fatal("expected correlation id header to be set")
	}
}

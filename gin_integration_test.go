package decimal

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGinBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should bind query", func(t *testing.T) {
		router := gin.New()
		router.GET("/query", func(c *gin.Context) {
			var req struct {
				Amount Decimal `form:"amount"`
			}

			if err := c.ShouldBindQuery(&req); err != nil {
				c.String(http.StatusBadRequest, err.Error())
				return
			}
			c.String(http.StatusOK, req.Amount.StringWithTrailingZeros())
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/query?amount=1.2300", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("ShouldBindQuery status = %d, want 200; body=%s", w.Code, w.Body.String())
		}
		if got := strings.TrimSpace(w.Body.String()); got != "1.2300" {
			t.Fatalf("ShouldBindQuery amount = %s, want 1.2300", got)
		}
	})

	t.Run("should bind uri", func(t *testing.T) {
		router := gin.New()
		router.GET("/orders/:amount", func(c *gin.Context) {
			var req struct {
				Amount Decimal `uri:"amount"`
			}

			if err := c.ShouldBindUri(&req); err != nil {
				c.String(http.StatusBadRequest, err.Error())
				return
			}
			c.String(http.StatusOK, req.Amount.StringWithTrailingZeros())
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/orders/7.5000", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("ShouldBindUri status = %d, want 200; body=%s", w.Code, w.Body.String())
		}
		if got := strings.TrimSpace(w.Body.String()); got != "7.5000" {
			t.Fatalf("ShouldBindUri amount = %s, want 7.5000", got)
		}
	})

	t.Run("should bind json", func(t *testing.T) {
		router := gin.New()
		router.POST("/json", func(c *gin.Context) {
			var req struct {
				Amount Decimal `json:"amount"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.String(http.StatusBadRequest, err.Error())
				return
			}
			c.String(http.StatusOK, req.Amount.StringWithTrailingZeros())
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/json", strings.NewReader(`{"amount":"3.1400"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("ShouldBindJSON status = %d, want 200; body=%s", w.Code, w.Body.String())
		}
		if got := strings.TrimSpace(w.Body.String()); got != "3.1400" {
			t.Fatalf("ShouldBindJSON amount = %s, want 3.1400", got)
		}
	})

	t.Run("should bind query invalid", func(t *testing.T) {
		router := gin.New()
		router.GET("/query", func(c *gin.Context) {
			var req struct {
				Amount Decimal `form:"amount"`
			}

			if err := c.ShouldBindQuery(&req); err != nil {
				c.String(http.StatusBadRequest, err.Error())
				return
			}
			c.String(http.StatusOK, req.Amount.String())
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/query?amount=not-a-decimal", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ShouldBindQuery(invalid) status = %d, want 400; body=%s", w.Code, w.Body.String())
		}
	})
}

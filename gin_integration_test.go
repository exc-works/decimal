package decimal

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func TestGinBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	trans := mustGetENTranslator(t)
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := RegisterGoPlaygroundValidator(v); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidator() returned error: %v", err)
		}
		if err := RegisterGoPlaygroundValidatorTranslations(v, trans); err != nil {
			t.Fatalf("RegisterGoPlaygroundValidatorTranslations() returned error: %v", err)
		}
	} else {
		t.Fatal("gin binding validator engine is not *validator.Validate")
	}

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

	t.Run("should validate required gt lte on query", func(t *testing.T) {
		router := gin.New()
		router.GET("/validate", func(c *gin.Context) {
			var req struct {
				Amount Decimal `form:"amount" binding:"decimal_required,decimal_gt=0,decimal_lte=10"`
			}

			if err := c.ShouldBindQuery(&req); err != nil {
				c.String(http.StatusBadRequest, err.Error())
				return
			}
			c.String(http.StatusOK, req.Amount.StringWithTrailingZeros())
		})

		cases := []struct {
			name       string
			url        string
			wantStatus int
		}{
			{name: "valid", url: "/validate?amount=1.2300", wantStatus: http.StatusOK},
			{name: "missing required", url: "/validate", wantStatus: http.StatusBadRequest},
			{name: "zero fails required", url: "/validate?amount=0", wantStatus: http.StatusBadRequest},
			{name: "negative fails gt", url: "/validate?amount=-1", wantStatus: http.StatusBadRequest},
			{name: "above lte fails", url: "/validate?amount=10.1", wantStatus: http.StatusBadRequest},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, tc.url, nil)
				router.ServeHTTP(w, req)

				if w.Code != tc.wantStatus {
					t.Fatalf("%s status = %d, want %d; body=%s", tc.name, w.Code, tc.wantStatus, w.Body.String())
				}
			})
		}
	})

	t.Run("should validate omitempty gte lt on query", func(t *testing.T) {
		router := gin.New()
		router.GET("/optional", func(c *gin.Context) {
			var req struct {
				Amount Decimal `form:"amount" binding:"omitempty,decimal_gte=1.5,decimal_lt=5"`
			}

			if err := c.ShouldBindQuery(&req); err != nil {
				c.String(http.StatusBadRequest, err.Error())
				return
			}
			c.String(http.StatusOK, req.Amount.StringWithTrailingZeros())
		})

		cases := []struct {
			name       string
			url        string
			wantStatus int
		}{
			{name: "missing is empty", url: "/optional", wantStatus: http.StatusOK},
			{name: "zero fails gte", url: "/optional?amount=0", wantStatus: http.StatusBadRequest},
			{name: "below gte fails", url: "/optional?amount=1.49", wantStatus: http.StatusBadRequest},
			{name: "valid middle", url: "/optional?amount=4.2", wantStatus: http.StatusOK},
			{name: "equal lt bound fails", url: "/optional?amount=5", wantStatus: http.StatusBadRequest},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, tc.url, nil)
				router.ServeHTTP(w, req)

				if w.Code != tc.wantStatus {
					t.Fatalf("%s status = %d, want %d; body=%s", tc.name, w.Code, tc.wantStatus, w.Body.String())
				}
			})
		}
	})

	t.Run("should validate eq on query", func(t *testing.T) {
		router := gin.New()
		router.GET("/eq", func(c *gin.Context) {
			var req struct {
				Amount Decimal `form:"amount" binding:"decimal_required,decimal_eq=1.23"`
			}

			if err := c.ShouldBindQuery(&req); err != nil {
				c.String(http.StatusBadRequest, err.Error())
				return
			}
			c.String(http.StatusOK, req.Amount.StringWithTrailingZeros())
		})

		cases := []struct {
			name       string
			url        string
			wantStatus int
		}{
			{name: "missing required", url: "/eq", wantStatus: http.StatusBadRequest},
			{name: "equal with same scale", url: "/eq?amount=1.23", wantStatus: http.StatusOK},
			{name: "equal with different scale", url: "/eq?amount=1.2300", wantStatus: http.StatusOK},
			{name: "not equal", url: "/eq?amount=1.24", wantStatus: http.StatusBadRequest},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, tc.url, nil)
				router.ServeHTTP(w, req)

				if w.Code != tc.wantStatus {
					t.Fatalf("%s status = %d, want %d; body=%s", tc.name, w.Code, tc.wantStatus, w.Body.String())
				}
			})
		}
	})

	t.Run("should return friendly validator message", func(t *testing.T) {
		router := gin.New()
		router.GET("/friendly", func(c *gin.Context) {
			var req struct {
				Amount Decimal `form:"amount" binding:"decimal_required,decimal_eq=1.23"`
			}

			if err := c.ShouldBindQuery(&req); err != nil {
				messages := TranslateGoPlaygroundValidationErrors(err, trans)
				c.String(http.StatusBadRequest, strings.Join(messages, "; "))
				return
			}
			c.String(http.StatusOK, req.Amount.StringWithTrailingZeros())
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/friendly", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("friendly status = %d, want 400; body=%s", w.Code, w.Body.String())
		}
		if got := strings.TrimSpace(w.Body.String()); got != "Amount is required" {
			t.Fatalf("friendly message = %q, want %q", got, "Amount is required")
		}
	})
}

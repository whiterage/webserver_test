package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		apiKey         string
		headerKey      string
		headerValue    string
		expectedStatus int
	}{
		{
			name:           "valid Bearer token",
			apiKey:         "secret-key",
			headerKey:      "Authorization",
			headerValue:    "Bearer secret-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid X-API-Key header",
			apiKey:         "secret-key",
			headerKey:      "X-API-Key",
			headerValue:    "secret-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid key",
			apiKey:         "secret-key",
			headerKey:      "X-API-Key",
			headerValue:    "wrong-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing key",
			apiKey:         "secret-key",
			headerKey:      "",
			headerValue:    "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(APIKeyAuth(tt.apiKey))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.headerKey != "" {
				req.Header.Set(tt.headerKey, tt.headerValue)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"mykanban-backend/auth"
	"mykanban-backend/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	router := gin.New()
	router.Use(JWTAuth("secret"))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Error == "" {
		t.Error("expected error message")
	}
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	router := gin.New()
	router.Use(JWTAuth("secret"))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidTokenFormat")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	router := gin.New()
	router.Use(JWTAuth("secret"))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuth_ValidToken(t *testing.T) {
	secret := "test-secret"
	token, _ := auth.GenerateJWT("test@example.com", secret, 1)

	router := gin.New()
	router.Use(JWTAuth(secret))
	router.GET("/test", func(c *gin.Context) {
		email, _ := c.Get("email")
		c.JSON(200, gin.H{"email": email})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["email"] != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %v", resp["email"])
	}
}

func TestRecovery_PanicHandled(t *testing.T) {
	router := gin.New()
	router.Use(Recovery())
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Error != "internal server error" {
		t.Errorf("expected 'internal server error', got '%s'", resp.Error)
	}
}

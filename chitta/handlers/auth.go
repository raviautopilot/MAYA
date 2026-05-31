package handlers

import (
	"chitta/auth"
	"chitta/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// ---- Auth Handlers ----

// Login authenticates with email/password and returns a JWT.
// @Summary      User login
// @Description  Authenticate with email and password to receive a JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body models.LoginRequest true "Login credentials"
// @Success      200 {object} models.APIResponse{data=object{token=string}} "JWT token"
// @Failure      400 {object} models.APIResponse "Invalid request"
// @Failure      401 {object} models.APIResponse "Invalid credentials"
// @Failure      500 {object} models.APIResponse "Token generation failure"
// @Router       /v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, nil, "invalid request: "+err.Error())
		return
	}
	if req.Email != h.AppConfig.RootEmail {
		respond(c, http.StatusUnauthorized, nil, "invalid credentials")
		return
	}
	if err := auth.CheckPassword(h.AppConfig.RootPasswordHash, req.Password); err != nil {
		respond(c, http.StatusUnauthorized, nil, "invalid credentials")
		return
	}
	token, err := auth.GenerateJWT(req.Email, h.AppConfig.JWTSecret, h.AppConfig.JWTExpiryHours)
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to generate token")
		return
	}
	respond(c, http.StatusOK, gin.H{"token": token}, "")
}

// ChangePassword updates the root password.
// @Summary      Change password
// @Description  Change the root user's password (requires current password)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body models.ChangePasswordRequest true "Password change request"
// @Success      200 {object} models.APIResponse{data=object{message=string}} "Password changed"
// @Failure      400 {object} models.APIResponse "Invalid request"
// @Failure      401 {object} models.APIResponse "Old password incorrect"
// @Failure      500 {object} models.APIResponse "Server error"
// @Router       /v1/auth/change-password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, nil, "invalid request: "+err.Error())
		return
	}
	if err := auth.CheckPassword(h.AppConfig.RootPasswordHash, req.OldPassword); err != nil {
		respond(c, http.StatusUnauthorized, nil, "old password is incorrect")
		return
	}
	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to hash new password")
		return
	}
	h.AppConfig.RootPasswordHash = newHash
	if err := h.Config.Save(h.AppConfig); err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to save config")
		return
	}
	respond(c, http.StatusOK, gin.H{"message": "password changed successfully"}, "")
}

// GoogleLogin redirects to Google OAuth consent screen.
// @Summary      Google OAuth login
// @Description  Redirect to Google OAuth 2.0 consent screen for authentication
// @Tags         Auth
// @Produce      json
// @Success      307 "Redirect to Google"
// @Failure      503 {object} models.APIResponse "Google OAuth not configured"
// @Router       /v1/auth/google/login [get]
func (h *Handler) GoogleLogin(c *gin.Context) {
	if h.AppConfig.DisableGoogleAuth || h.AppConfig.GoogleClientID == "" {
		respond(c, http.StatusServiceUnavailable, nil, "google oauth not configured or disabled")
		return
	}
	if h.AppConfig.GoogleClientID == "mock" {
		c.Redirect(http.StatusTemporaryRedirect, "/api/v1/auth/google/mock-consent")
		return
	}
	cfg := auth.GoogleOAuthConfig(h.AppConfig.GoogleClientID, h.AppConfig.GoogleClientSecret, h.AppConfig.GoogleRedirectURL)
	url := cfg.AuthCodeURL("state-token")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// MockConsent serves a mock Google consent HTML page for E2E testing
func (h *Handler) MockConsent(c *gin.Context) {
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <title>Mock Google Consent Portal</title>
</head>
<body style="background-color: #f3f4f6; margin: 0; padding: 0;">
    <div style="text-align: center; margin-top: 100px; font-family: sans-serif; max-width: 400px; margin-left: auto; margin-right: auto; background: white; padding: 40px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
        <h2 style="color: #1f2937; margin-bottom: 10px;">Mock Google Consent Portal</h2>
        <p style="color: #6b7280; font-size: 14px; margin-bottom: 30px;">This is a simulated Google authorization screen for local testing.</p>
        <a href="/api/v1/auth/google/callback?code=mock-code-123" data-testid="mock-login-user-btn" style="display: inline-block; padding: 12px 24px; background-color: #3b82f6; color: white; text-decoration: none; border-radius: 6px; font-weight: 600; font-size: 14px; transition: background-color 0.2s;">
            Authorize Mock Profile
        </a>
    </div>
</body>
</html>
`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
}

// GoogleCallback handles the OAuth callback from Google.
// @Summary      Google OAuth callback
// @Description  Handle the callback from Google OAuth and return a JWT token
// @Tags         Auth
// @Produce      json
// @Param        code query string true "Authorization code from Google"
// @Success      200 {object} models.APIResponse{data=object{token=string,email=string,name=string}} "JWT token and user info"
// @Failure      400 {object} models.APIResponse "Missing code parameter"
// @Failure      500 {object} models.APIResponse "Google auth or token generation failure"
// @Router       /v1/auth/google/callback [get]
func (h *Handler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		respond(c, http.StatusBadRequest, nil, "missing code parameter")
		return
	}

	var email string
	var name string

	if h.AppConfig.GoogleClientID == "mock" && code == "mock-code-123" {
		email = "developer@mykanban.local"
		name = "Test Developer"
	} else {
		cfg := auth.GoogleOAuthConfig(h.AppConfig.GoogleClientID, h.AppConfig.GoogleClientSecret, h.AppConfig.GoogleRedirectURL)
		userInfo, err := auth.FetchGoogleUser(c.Request.Context(), cfg, code)
		if err != nil {
			respond(c, http.StatusInternalServerError, nil, "google auth failed: "+err.Error())
			return
		}
		email = userInfo.Email
		name = userInfo.Name
	}

	token, err := auth.GenerateJWT(email, h.AppConfig.JWTSecret, h.AppConfig.JWTExpiryHours)
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to generate token")
		return
	}

	// Always redirect back to frontend login callback page with tokens for browser E2E flows
	frontendURL := "http://localhost:3000"
	if h.AppConfig.AllowedOrigins != "" {
		origins := parseAllowedOrigins(h.AppConfig.AllowedOrigins)
		if len(origins) > 0 {
			frontendURL = origins[0]
		}
	}

	redirectURL := fmt.Sprintf("%s/login?token=%s&email=%s&name=%s", frontendURL, token, email, name)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func parseAllowedOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{"http://localhost:3000"}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	if len(origins) == 0 {
		return []string{"http://localhost:3000"}
	}
	return origins
}

// GetAuthConfig returns public authentication settings
func (h *Handler) GetAuthConfig(c *gin.Context) {
	googleAuthEnabled := !h.AppConfig.DisableGoogleAuth && h.AppConfig.GoogleClientID != ""
	respond(c, http.StatusOK, gin.H{"google_auth_enabled": googleAuthEnabled}, "")
}

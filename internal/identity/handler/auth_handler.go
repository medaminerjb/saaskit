// Package handler provides HTTP handlers for the identity module.
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/medaminerjb/saas-kit/internal/identity/domain"
	"github.com/medaminerjb/saas-kit/internal/identity/service"
)

// AuthHandler handles authentication HTTP endpoints.
type AuthHandler struct {
	identity *service.IdentityManager
	logger   *slog.Logger
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(identity *service.IdentityManager, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{identity: identity, logger: logger}
}

// Routes registers auth routes on the given router.
func (h *AuthHandler) Routes(r chi.Router) {
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.RefreshTokens)
	r.Post("/auth/forgot-password", h.ForgotPassword)
	r.Post("/auth/reset-password", h.ResetPassword)
	r.Post("/auth/verify-email", h.VerifyEmail)
}

// ProtectedRoutes registers auth routes that require authentication.
func (h *AuthHandler) ProtectedRoutes(r chi.Router) {
	r.Post("/auth/logout", h.Logout)
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// Register handles POST /api/v1/auth/register
// @Summary Register a new user
// @Description Create a new user account with email, password, and name
// @Tags auth
// @Accept json
// @Produce json
// @Param request body registerRequest true "Registration details"
// @Success 201 {object} map[string]interface{} "User created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 409 {object} map[string]string "User already exists"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, tokens, err := h.identity.Register(r.Context(), service.RegisterInput{
		Email:    strings.TrimSpace(req.Email),
		Password: req.Password,
		Name:     strings.TrimSpace(req.Name),
	})
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login handles POST /api/v1/auth/login
// @Summary User login
// @Description Authenticate a user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 403 {object} map[string]string "Account disabled or locked"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, tokens, err := h.identity.Login(r.Context(), service.LoginInput{
		Email:     strings.TrimSpace(req.Email),
		Password:  req.Password,
		UserAgent: r.UserAgent(),
		IPAddress: extractIP(r),
	})
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokens handles POST /api/v1/auth/refresh
// @Summary Refresh access token
// @Description Get a new access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body refreshRequest true "Refresh token"
// @Success 200 {object} map[string]interface{} "Tokens refreshed"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid or expired token"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tokens, err := h.identity.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

// Logout handles POST /api/v1/auth/logout
// @Summary User logout
// @Description Invalidate the current session
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]string "Logout successful"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal error"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	sessionID, err := uuid.Parse(claims.SessionID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user")
		return
	}

	if err := h.identity.Logout(r.Context(), sessionID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "logout failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

// ForgotPassword handles POST /api/v1/auth/forgot-password
// @Summary Request password reset
// @Description Send a password reset email to the user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body forgotPasswordRequest true "Email address"
// @Success 200 {object} map[string]string "Reset email sent (if account exists)"
// @Failure 400 {object} map[string]string "Invalid request"
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Always return success to prevent email enumeration
	_ = h.identity.Auth.RequestPasswordReset(r.Context(), strings.TrimSpace(req.Email))

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "if an account with that email exists, a reset link has been sent",
	})
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ResetPassword handles POST /api/v1/auth/reset-password
// @Summary Reset password
// @Description Reset password using a token from email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body resetPasswordRequest true "Reset token and new password"
// @Success 200 {object} map[string]string "Password reset successful"
// @Failure 400 {object} map[string]string "Invalid request or token"
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.identity.Auth.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		h.handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password reset successfully"})
}

type verifyEmailRequest struct {
	Token string `json:"token"`
}

// VerifyEmail handles POST /api/v1/auth/verify-email
// @Summary Verify email address
// @Description Verify user email using a token from email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body verifyEmailRequest true "Verification token"
// @Success 200 {object} map[string]string "Email verified"
// @Failure 400 {object} map[string]string "Invalid request or token"
// @Router /auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.identity.Auth.VerifyEmail(r.Context(), req.Token); err != nil {
		h.handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "email verified"})
}

func (h *AuthHandler) handleAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, domain.ErrAccountDisabled):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, domain.ErrAccountLocked):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, domain.ErrInvalidToken):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, domain.ErrSessionExpired):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, domain.ErrInvalidEmail):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrPasswordTooShort),
		errors.Is(err, domain.ErrPasswordTooLong),
		errors.Is(err, domain.ErrPasswordRequired):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Error("auth error", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/medaminerjb/saas-kit/internal/identity/service"
)

// UserHandler handles user profile HTTP endpoints.
type UserHandler struct {
	identity *service.IdentityManager
	logger   *slog.Logger
}

// NewUserHandler creates a new user handler.
func NewUserHandler(identity *service.IdentityManager, logger *slog.Logger) *UserHandler {
	return &UserHandler{identity: identity, logger: logger}
}

// Routes registers user routes (all require authentication).
func (h *UserHandler) Routes(r chi.Router) {
	r.Get("/users/me", h.GetMe)
	r.Patch("/users/me", h.UpdateMe)
	r.Get("/users/me/sessions", h.ListSessions)
	r.Delete("/users/me/sessions/{sessionID}", h.RevokeSession)
	r.Get("/users/me/metadata", h.GetMetadata)
	r.Patch("/users/me/metadata", h.UpdateMetadata)
}

// GetMe handles GET /api/v1/users/me
// @Summary Get current user
// @Description Get the authenticated user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "User profile"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Router /users/me [get]
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.identity.GetCurrentUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

type updateMeRequest struct {
	Name      *string `json:"name,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// UpdateMe handles PATCH /api/v1/users/me
// @Summary Update current user
// @Description Update the authenticated user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body updateMeRequest true "Profile updates"
// @Success 200 {object} map[string]interface{} "Updated user profile"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Update failed"
// @Router /users/me [patch]
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req updateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.identity.UpdateProfile(r.Context(), userID, service.UpdateProfileInput{
		Name:      req.Name,
		AvatarURL: req.AvatarURL,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "update failed")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// ListSessions handles GET /api/v1/users/me/sessions
// @Summary List user sessions
// @Description Get all active sessions for the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "List of sessions"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /users/me/sessions [get]
func (h *UserHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Stub: will be implemented when session listing is wired through the IdentityManager
	writeJSON(w, http.StatusOK, map[string]any{"sessions": []any{}})
}

// RevokeSession handles DELETE /api/v1/users/me/sessions/{sessionID}
// @Summary Revoke a session
// @Description Revoke a specific session for the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param sessionID path string true "Session ID"
// @Success 200 {object} map[string]string "Session revoked"
// @Failure 400 {object} map[string]string "Invalid session ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Revoke failed"
// @Router /users/me/sessions/{sessionID} [delete]
func (h *UserHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := h.identity.Logout(r.Context(), sessionID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "revoke failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "session revoked"})
}

// GetMetadata handles GET /api/v1/users/me/metadata
// @Summary Get user metadata
// @Description Get public metadata for the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "User metadata"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Router /users/me/metadata [get]
func (h *UserHandler) GetMetadata(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.identity.GetCurrentUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"metadata_public": user.MetadataPublic,
	})
}

type updateMetadataRequest struct {
	MetadataPublic  map[string]interface{} `json:"metadata_public,omitempty"`
	MetadataPrivate map[string]interface{} `json:"metadata_private,omitempty"`
}

// UpdateMetadata handles PATCH /api/v1/users/me/metadata
// @Summary Update user metadata
// @Description Update metadata for the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body updateMetadataRequest true "Metadata updates"
// @Success 200 {object} map[string]interface{} "Updated metadata"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Update failed"
// @Router /users/me/metadata [patch]
func (h *UserHandler) UpdateMetadata(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req updateMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.identity.UpdateUserMetadata(r.Context(), userID, service.UpdateMetadataInput{
		MetadataPublic:  req.MetadataPublic,
		MetadataPrivate: req.MetadataPrivate,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "update failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"metadata_public": user.MetadataPublic,
	})
}

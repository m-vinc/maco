package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/m-vinc/maco/pkg/types"
)

type CreateAPIKeyRequest struct {
	Name string `json:"name"`
}

type CreateAPIKeyResponse struct {
	APIKey types.APIKey `json:"api_key"`
	Token  string       `json:"token"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

// @Summary listAPIKeys
// @ID listAPIKeys
// @Tags api-keys
// @Security BearerAuth
// @Produce json
// @Success 200 {array} types.APIKey
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/me/api-keys [get]
func (s *Server) listAPIKeys(w http.ResponseWriter, r *http.Request) {
	claims := claimsFrom(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	keys, err := s.engine.ListAPIKeys(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, keys)
}

// @Summary createAPIKey
// @ID createAPIKey
// @Tags api-keys
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body CreateAPIKeyRequest true "Request"
// @Success 201 {object} CreateAPIKeyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/me/api-keys [post]
func (s *Server) createAPIKey(w http.ResponseWriter, r *http.Request) {
	claims := claimsFrom(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateAPIKeyRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	key, token, err := s.engine.CreateAPIKey(r.Context(), claims.UserID, req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, CreateAPIKeyResponse{APIKey: key, Token: token})
}

// @Summary revokeAPIKey
// @ID revokeAPIKey
// @Tags api-keys
// @Security BearerAuth
// @Produce json
// @Param id path string true "API key id"
// @Success 200 {object} StatusResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/me/api-keys/{id} [delete]
func (s *Server) revokeAPIKey(w http.ResponseWriter, r *http.Request) {
	claims := claimsFrom(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	revoked, err := s.engine.RevokeAPIKey(r.Context(), claims.UserID, chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !revoked {
		writeError(w, http.StatusNotFound, "API key not found")
		return
	}

	writeJSON(w, http.StatusOK, StatusResponse{Status: "ok"})
}

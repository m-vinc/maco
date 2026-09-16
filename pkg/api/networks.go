package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/m-vinc/maco/pkg/engine"
	"github.com/m-vinc/maco/pkg/jobs"
)

// @Summary listNetworks
// @ID listNetworks
// @Tags networks
// @Security BearerAuth
// @Produce json
// @Success 200 {array} types.NetworkManifest
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/networks [get]
func (s *Server) listNetworks(w http.ResponseWriter, r *http.Request) {
	networks, err := s.engine.ListNetworks()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, networks)
}

// @Summary createNetwork
// @ID createNetwork
// @Tags networks
// @Security BearerAuth
// @Produce json
// @Accept json
// @Param body body engine.CreateNetworkParams true "Request"
// @Success 202 {object} jobs.Job
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/networks [post]
func (s *Server) createNetwork(w http.ResponseWriter, r *http.Request) {
	var params engine.CreateNetworkParams
	if err := readJSON(r, &params); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	s.submitJob(w, r, jobs.Payload{Action: "network.create", Target: params.Name, Network: params})
}

// @Summary applyNetwork
// @ID applyNetwork
// @Tags networks
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Success 202 {object} jobs.Job
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/networks/{id}/apply [post]
func (s *Server) applyNetwork(w http.ResponseWriter, r *http.Request) {
	s.submitJob(w, r, jobs.Payload{Action: "network.apply", Target: chi.URLParam(r, "id")})
}

// @Summary destroyNetwork
// @ID destroyNetwork
// @Tags networks
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Success 202 {object} jobs.Job
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/networks/{id} [delete]
func (s *Server) destroyNetwork(w http.ResponseWriter, r *http.Request) {
	s.submitJob(w, r, jobs.Payload{Action: "network.destroy", Target: chi.URLParam(r, "id")})
}

// @Summary updateNetwork
// @ID updateNetwork
// @Tags networks
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Accept json
// @Param body body engine.CreateNetworkParams true "Request"
// @Success 202 {object} jobs.Job
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/networks/{id} [put]
func (s *Server) updateNetwork(w http.ResponseWriter, r *http.Request) {
	var params engine.CreateNetworkParams
	if err := readJSON(r, &params); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	s.submitJob(w, r, jobs.Payload{Action: "network.update", Target: chi.URLParam(r, "id"), Network: params})
}

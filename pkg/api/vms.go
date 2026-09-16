package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/m-vinc/maco/pkg/engine"
	"github.com/m-vinc/maco/pkg/jobs"
)

// @Summary listVMs
// @ID listVMs
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Success 200 {array} engine.VMView
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/vms [get]
func (s *Server) listVMs(w http.ResponseWriter, r *http.Request) {
	views, err := s.engine.ListVMs(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, views)
}

// @Summary createVM
// @ID createVM
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Accept json
// @Param body body engine.CreateVMParams true "Request"
// @Success 202 {object} jobs.Job
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms [post]
func (s *Server) createVM(w http.ResponseWriter, r *http.Request) {
	var params engine.CreateVMParams
	if err := readJSON(r, &params); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	s.submitJob(w, r, jobs.Payload{Action: "vm.create", Target: params.Name, VM: params})
}

// @Summary getVM
// @ID getVM
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} engine.VMView
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/vms/{id} [get]
func (s *Server) getVM(w http.ResponseWriter, r *http.Request) {
	m, st, err := s.engine.StatusVM(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	views, err := s.engine.ListVMs(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, view := range views {
		if view.Manifest.ID == m.ID {
			writeJSON(w, http.StatusOK, view)
			return
		}
	}

	writeJSON(w, http.StatusOK, engine.VMView{Manifest: m, Phase: string(st.Phase), PID: st.PID})
}

// @Summary startVM
// @ID startVM
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Success 202 {object} jobs.Job
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/start [post]
func (s *Server) startVM(w http.ResponseWriter, r *http.Request) {
	s.submitJob(w, r, jobs.Payload{Action: "vm.start", Target: chi.URLParam(r, "id")})
}

// @Summary stopVM
// @ID stopVM
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Success 202 {object} jobs.Job
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/stop [post]
func (s *Server) stopVM(w http.ResponseWriter, r *http.Request) {
	s.submitJob(w, r, jobs.Payload{Action: "vm.stop", Target: chi.URLParam(r, "id")})
}

// @Summary deleteVM
// @ID deleteVM
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Success 202 {object} jobs.Job
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id} [delete]
func (s *Server) deleteVM(w http.ResponseWriter, r *http.Request) {
	s.submitJob(w, r, jobs.Payload{Action: "vm.delete", Target: chi.URLParam(r, "id")})
}

// @Summary previewVM
// @ID previewVM
// @Tags vms
// @Security BearerAuth
// @Produce png
// @Param id path string true "id"
// @Success 200 {file} binary "PNG preview"
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/vms/{id}/preview [get]
func (s *Server) previewVM(w http.ResponseWriter, r *http.Request) {
	manifest, _, err := s.engine.StatusVM(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	file, err := os.Open(filepath.Join(s.engine.Paths().VMRunDir(manifest.ID), "preview.png"))
	if err != nil {
		writeError(w, http.StatusNotFound, "preview unavailable")
		return
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Preview-Captured-At", strconv.FormatInt(info.ModTime().UnixMilli(), 10))
	w.Header().Set("Content-Type", "image/png")
	http.ServeContent(w, r, "preview.png", info.ModTime(), file)
}

// @Summary updateHardware
// @ID updateHardware
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Accept json
// @Param body body engine.UpdateHardwareParams true "Request"
// @Success 202 {object} jobs.Job
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/hardware [patch]
func (s *Server) updateHardware(w http.ResponseWriter, r *http.Request) {
	var params engine.UpdateHardwareParams
	if err := readJSON(r, &params); err != nil {
		writeError(w, http.StatusBadRequest, "invalid hardware settings")
		return
	}

	s.submitJob(w, r, jobs.Payload{Action: "vm.hardware", Target: chi.URLParam(r, "id"), Hardware: params})
}

// @Summary shutdownVM
// @ID shutdownVM
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Success 202 {object} jobs.Job
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/shutdown [post]
func (s *Server) shutdownVM(w http.ResponseWriter, r *http.Request) {
	s.submitJob(w, r, jobs.Payload{Action: "vm.shutdown", Target: chi.URLParam(r, "id")})
}

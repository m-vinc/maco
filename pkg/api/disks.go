package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/m-vinc/maco/pkg/engine"
	"github.com/m-vinc/maco/pkg/jobs"
)

// @Summary listDisks
// @ID listDisks
// @Tags disks
// @Security BearerAuth
// @Produce json
// @Param page query int false "page"
// @Param page_size query int false "page_size"
// @Success 200 {object} api.Page[engine.DiskView]
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/disks [get]
func (s *Server) listDisks(w http.ResponseWriter, r *http.Request) {
	page, pageSize, ok := pageParams(w, r)
	if !ok {
		return
	}

	disks, err := s.engine.ListDisks()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, paginate(disks, page, pageSize))
}

// @Summary diskStorage
// @ID diskStorage
// @Tags disks
// @Security BearerAuth
// @Produce json
// @Success 200 {object} engine.StorageStats
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/disks/stats [get]
func (s *Server) diskStorage(w http.ResponseWriter, r *http.Request) {
	stats, err := s.engine.StorageStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// @Summary addDisk
// @ID addDisk
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Accept json
// @Param body body engine.DiskParams true "Request"
// @Success 202 {object} jobs.Job
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/disks [post]
func (s *Server) addDisk(w http.ResponseWriter, r *http.Request) {
	var params engine.DiskParams
	if err := readJSON(r, &params); err != nil {
		writeError(w, http.StatusBadRequest, "invalid disk settings")
		return
	}

	s.submitJob(w, r, jobs.Payload{Action: "vm.disk.add", Target: chi.URLParam(r, "id"), Disk: params})
}

// @Summary growDisk
// @ID growDisk
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Param disk path string true "disk"
// @Accept json
// @Param body body engine.DiskParams true "Request"
// @Success 202 {object} jobs.Job
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/disks/{disk} [patch]
func (s *Server) growDisk(w http.ResponseWriter, r *http.Request) {
	var params engine.DiskParams
	if err := readJSON(r, &params); err != nil {
		writeError(w, http.StatusBadRequest, "invalid disk settings")
		return
	}

	params.ID = chi.URLParam(r, "disk")
	s.submitJob(w, r, jobs.Payload{Action: "vm.disk.grow", Target: chi.URLParam(r, "id"), Disk: params})
}

// @Summary removeDisk
// @ID removeDisk
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Param disk path string true "disk"
// @Success 202 {object} jobs.Job
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/disks/{disk} [delete]
func (s *Server) removeDisk(w http.ResponseWriter, r *http.Request) {
	params := engine.DiskParams{ID: chi.URLParam(r, "disk")}
	s.submitJob(w, r, jobs.Payload{Action: "vm.disk.remove", Target: chi.URLParam(r, "id"), Disk: params})
}

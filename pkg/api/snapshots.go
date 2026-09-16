package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/m-vinc/maco/pkg/engine"
	"github.com/m-vinc/maco/pkg/jobs"
)

// @Summary listSnapshots
// @ID listSnapshots
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Success 200 {array} vm.Snapshot
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/vms/{id}/snapshots [get]
func (s *Server) listSnapshots(w http.ResponseWriter, r *http.Request) {
	snapshots, err := s.engine.ListSnapshots(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, snapshots)
}

// @Summary createSnapshot
// @ID createSnapshot
// @Tags vms
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body engine.SnapshotParams true "Request"
// @Success 202 {object} jobs.Job
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/snapshots [post]
func (s *Server) createSnapshot(w http.ResponseWriter, r *http.Request) {
	var params engine.SnapshotParams
	if err := readJSON(r, &params); err != nil {
		writeError(w, http.StatusBadRequest, "invalid snapshot settings")
		return
	}

	s.submitJob(w, r, jobs.Payload{Action: "vm.snapshot.create", Target: chi.URLParam(r, "id"), Snapshot: params})
}

// @Summary restoreSnapshot
// @ID restoreSnapshot
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Param tag path string true "tag"
// @Success 202 {object} jobs.Job
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/snapshots/{tag}/restore [post]
func (s *Server) restoreSnapshot(w http.ResponseWriter, r *http.Request) {
	s.submitJob(w, r, jobs.Payload{
		Action:   "vm.snapshot.restore",
		Target:   chi.URLParam(r, "id"),
		Snapshot: engine.SnapshotParams{Tag: chi.URLParam(r, "tag")},
	})
}

// @Summary deleteSnapshot
// @ID deleteSnapshot
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Param tag path string true "tag"
// @Success 202 {object} jobs.Job
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/snapshots/{tag} [delete]
func (s *Server) deleteSnapshot(w http.ResponseWriter, r *http.Request) {
	s.submitJob(w, r, jobs.Payload{
		Action:   "vm.snapshot.delete",
		Target:   chi.URLParam(r, "id"),
		Snapshot: engine.SnapshotParams{Tag: chi.URLParam(r, "tag")},
	})
}

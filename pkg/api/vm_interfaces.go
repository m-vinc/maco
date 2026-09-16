package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/m-vinc/maco/pkg/engine"
	"github.com/m-vinc/maco/pkg/jobs"
	"net/http"
)

// @Summary addVMInterface
// @ID addVMInterface
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Accept json
// @Param body body engine.InterfaceParams true "Request"
// @Success 202 {object} jobs.Job
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/interfaces [post]
func (s *Server) addVMInterface(w http.ResponseWriter, r *http.Request) {
	s.submitInterface(w, r, "vm.interface.add")
}

// @Summary updateVMInterface
// @ID updateVMInterface
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Param interface path string true "interface"
// @Accept json
// @Param body body engine.InterfaceParams true "Request"
// @Success 202 {object} jobs.Job
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/interfaces/{interface} [patch]
func (s *Server) updateVMInterface(w http.ResponseWriter, r *http.Request) {
	s.submitInterface(w, r, "vm.interface.update")
}

// @Summary removeVMInterface
// @ID removeVMInterface
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Param interface path string true "interface"
// @Success 202 {object} jobs.Job
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/interfaces/{interface} [delete]
func (s *Server) removeVMInterface(w http.ResponseWriter, r *http.Request) {
	s.submitInterface(w, r, "vm.interface.remove")
}
func (s *Server) submitInterface(w http.ResponseWriter, r *http.Request, action string) {
	var p engine.InterfaceParams
	if action != "vm.interface.remove" {
		if err := readJSON(r, &p); err != nil {
			writeError(w, 400, "invalid interface settings")
			return
		}
	}
	p.ID = chi.URLParam(r, "interface")
	s.submitJob(w, r, jobs.Payload{Action: action, Target: chi.URLParam(r, "id"), Interface: p})
}

package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/m-vinc/maco/pkg/engine"
	"github.com/m-vinc/maco/pkg/jobs"
)

// @Summary listUSBDevices
// @ID listUSBDevices
// @Tags usb
// @Security BearerAuth
// @Produce json
// @Success 200 {object} engine.USBInventory
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/usb/devices [get]
func (s *Server) listUSBDevices(w http.ResponseWriter, r *http.Request) {
	result, err := s.engine.ListUSBDevices(r.Context())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// @Summary listVMUSB
// @ID listVMUSB
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Success 200 {array} engine.USBAttachment
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/vms/{id}/usb [get]
func (s *Server) listVMUSB(w http.ResponseWriter, r *http.Request) {
	result, err := s.engine.ListVMUSB(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// @Summary attachUSB
// @ID attachUSB
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Accept json
// @Param body body engine.USBParams true "Request"
// @Success 202 {object} jobs.Job
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/usb [post]
func (s *Server) attachUSB(w http.ResponseWriter, r *http.Request) {
	var params engine.USBParams
	if err := readJSON(r, &params); err != nil {
		writeError(w, http.StatusBadRequest, "invalid USB selection")
		return
	}
	if err := params.ValidateAttach(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	m, st, err := s.engine.StatusVM(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if st.Phase != "running" {
		writeError(w, http.StatusConflict, "start the VM before attaching USB devices")
		return
	}
	s.submitJob(w, r, jobs.Payload{Action: "vm.usb.attach", Target: m.ID, USB: params})
}

// @Summary detachUSB
// @ID detachUSB
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Param attachment path string true "attachment"
// @Success 202 {object} jobs.Job
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/usb/{attachment} [delete]
func (s *Server) detachUSB(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "attachment")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, http.StatusBadRequest, "invalid USB attachment ID")
		return
	}
	m, _, err := s.engine.StatusVM(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.submitJob(w, r, jobs.Payload{Action: "vm.usb.detach", Target: m.ID, USB: engine.USBParams{AttachmentID: id}})
}

// @Summary assignUSB
// @ID assignUSB
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Accept json
// @Param body body engine.USBParams true "Request"
// @Success 202 {object} jobs.Job
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/usb/assignments [post]
func (s *Server) assignUSB(w http.ResponseWriter, r *http.Request) {
	var params engine.USBParams
	if err := readJSON(r, &params); err != nil {
		writeError(w, http.StatusBadRequest, "invalid USB selection")
		return
	}
	if err := params.ValidateAttach(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	m, _, err := s.engine.StatusVM(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.submitJob(w, r, jobs.Payload{Action: "vm.usb.assign", Target: m.ID, USB: params})
}

// @Summary unassignUSB
// @ID unassignUSB
// @Tags vms
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Param key path string true "key"
// @Success 202 {object} jobs.Job
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/vms/{id}/usb/assignments/{key} [delete]
func (s *Server) unassignUSB(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" || len(key) > 512 {
		writeError(w, http.StatusBadRequest, "invalid USB assignment")
		return
	}
	m, _, err := s.engine.StatusVM(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.submitJob(w, r, jobs.Payload{Action: "vm.usb.unassign", Target: m.ID, USB: engine.USBParams{AssignmentKey: key}})
}

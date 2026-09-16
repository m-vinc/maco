package api

import "net/http"

// @Summary hostInfo
// @ID hostInfo
// @Tags host
// @Security BearerAuth
// @Produce json
// @Success 200 {object} engine.HostInfo
// @Failure 401 {object} ErrorResponse
// @Router /api/host [get]
func (s *Server) hostInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.engine.HostInfo())
}

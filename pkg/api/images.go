package api

import (
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/m-vinc/maco/pkg/image"
	"github.com/m-vinc/maco/pkg/jobs"
)

// @Summary listImages
// @ID listImages
// @Tags images
// @Security BearerAuth
// @Produce json
// @Success 200 {array} string
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/images [get]
func (s *Server) listImages(w http.ResponseWriter, r *http.Request) {
	names := make([]string, 0, len(image.Catalog))
	for name := range image.Catalog {
		names = append(names, name)
	}

	media, err := s.engine.ListMedia()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	for _, m := range media {
		if m.Kind == "image" {
			names = append(names, "media:"+m.ID)
		}
	}
	sort.Strings(names)
	writeJSON(w, http.StatusOK, names)
}

// @Summary listCatalog
// @ID listCatalog
// @Tags catalog
// @Security BearerAuth
// @Produce json
// @Success 200 {array} engine.CatalogImage
// @Failure 401 {object} ErrorResponse
// @Router /api/catalog [get]
func (s *Server) listCatalog(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.engine.Catalog())
}

// @Summary downloadCatalogImage
// @ID downloadCatalogImage
// @Tags catalog
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Success 202 {object} jobs.Job
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/catalog/{id}/download [post]
func (s *Server) downloadCatalogImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	s.submitJob(w, r, jobs.Payload{Action: "catalog.download", Target: id})
}

// @Summary deleteCatalogImage
// @ID deleteCatalogImage
// @Tags catalog
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Success 204 "No content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/catalog/{id} [delete]
func (s *Server) deleteCatalogImage(w http.ResponseWriter, r *http.Request) {
	if err := s.engine.DeleteCatalogImage(chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

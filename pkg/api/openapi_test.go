package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

type openAPIDocument struct {
	OpenAPI string                                `json:"openapi"`
	Paths   map[string]map[string]json.RawMessage `json:"paths"`
}

func TestOpenAPIRoutes(t *testing.T) {
	s := &Server{}
	s.router = s.buildRouter()
	response := httptest.NewRecorder()
	s.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/docs/openapi.json", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("spec status: %d", response.Code)
	}

	var document openAPIDocument
	if err := json.Unmarshal(response.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if document.OpenAPI != "3.1.0" {
		t.Fatalf("OpenAPI version: %s", document.OpenAPI)
	}

	mounted := map[string]bool{}
	err := chi.Walk(s.router, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if !strings.HasPrefix(route, "/api/") || route == "/api/docs/openapi.json" {
			return nil
		}
		key := strings.ToLower(method)
		mounted[key+" "+route] = true
		if _, ok := document.Paths[route][key]; !ok {
			t.Errorf("undocumented route: %s %s", method, route)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for route, methods := range document.Paths {
		for method := range methods {
			if !mounted[method+" "+route] {
				t.Errorf("unmounted operation: %s %s", method, route)
			}
		}
	}
}

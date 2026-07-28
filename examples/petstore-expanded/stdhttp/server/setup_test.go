package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewServerServesOpenAPIDocument(t *testing.T) {
	srv, err := NewServer()
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	srv.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/openapi.json", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("GET /openapi.json status = %d, want %d", response.Code, http.StatusOK)
	}
}

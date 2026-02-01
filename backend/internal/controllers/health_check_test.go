package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupHealthTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctrl := NewHealthCheckController()
	router.GET("/health", ctrl.Get)
	return router
}

func TestHealthCheck_Returns200OK(t *testing.T) {
	router := setupHealthTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHealthCheck_CorrectJSONBody(t *testing.T) {
	router := setupHealthTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", resp["status"])
	}
	if resp["message"] != "Backend is running" {
		t.Errorf("expected message 'Backend is running', got %q", resp["message"])
	}
}

func TestHealthCheck_ContentTypeHeader(t *testing.T) {
	router := setupHealthTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	expected := "application/json; charset=utf-8"
	if contentType != expected {
		t.Errorf("expected Content-Type %q, got %q", expected, contentType)
	}
}

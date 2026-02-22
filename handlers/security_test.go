package handlers

import (
	"errors"
	"firefly-importer/config"
	"firefly-importer/firefly"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRenderErrorSensitiveExposure(t *testing.T) {
	// Mock Firefly client
	client := firefly.NewClient("http://mock", "test-token")
	cfg := &config.Config{}
	h := NewAppHandler(client, cfg)

	rr := httptest.NewRecorder()
	sensitiveErr := errors.New("sensitive information: database_password=12345")

	h.renderError(rr, http.StatusInternalServerError, "An error occurred", sensitiveErr)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}

	body := rr.Body.String()
	if strings.Contains(body, "database_password=12345") {
		t.Errorf("sensitive information exposed in response: %s", body)
	}
}

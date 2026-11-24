package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Miroku77/goprxveil/internal/handlers"
)

func TestGetAnalyticsHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		expectedStatus int
	}{
		{
			name:           "GET not valid url",
			method:         http.MethodGet,
			url:            "/api/analytics/abc",
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			handlers.GetAnalyticsHandler(rr, req)
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected %d, actual %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

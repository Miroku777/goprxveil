package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Miroku77/goprxveil/internal/handlers"
)

func TestManipulateURLHandler_Validation(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		expectedStatus int
	}{
		{
			name:           "GET group_id nil",
			method:         http.MethodGet,
			url:            "/api/urls",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "GET group_id not valid",
			method:         http.MethodGet,
			url:            "/api/urls?group_id=abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "POST JSON incorrect",
			method:         http.MethodPost,
			url:            "/api/urls",
			body:           `{"group_id": 1, "url": `,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			handlers.ManipulateURLHandler(rr, req)
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected %d, actual %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

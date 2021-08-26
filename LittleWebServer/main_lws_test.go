package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLittleServer_ServeHTTP(t *testing.T) {
	t.Run("Do a get", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/api/people/123", nil)
		response := httptest.NewRecorder() // built-in mocks! :-)

		server := &LittleServer{}
		server.ServeHTTP(response, request)

		expected := `{"error":"page not found"}`
		actual := response.Body.String()

		if actual != expected {
			t.Errorf("Expected\r\n    %v\r\nBut got\r\n    %v\r\n", expected, actual)
		}
	})
}

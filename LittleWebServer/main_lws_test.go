package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestScanning(t *testing.T){

	bits := strings.Split("/x/y/z", "/")[1:]

	fmt.Printf("Got back %v\r\n", bits)
}

func TestLittleServer_ServeHTTP(t *testing.T) {
	t.Run("Do a get", func(t *testing.T) {
		SetUpLogging(false, false)
		request, _ := http.NewRequest(http.MethodGet, "http://localhost:6080/api/people/123", nil)
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

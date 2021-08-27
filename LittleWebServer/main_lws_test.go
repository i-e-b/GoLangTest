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

func TestPasswordProtection(t *testing.T){
	server :=  &LittleServer{userDb: map[int]MyInputType{}}
	server.SetUpLogging(false, false)
	server.userDb[0] = MyInputType{
		ID:   123,
		Name: "Test user",
		Age:  22,
	}

	expectedRejection := `{"error":"must provide token cookie"}`
	expectedSuccess := `{"id":123,"name":"Test user","age":22}`
	loginString := `{ "username": "ieb", "password": "correct" }`


	// First, should be rejected without a token
	request, _ := http.NewRequest(http.MethodGet, "http://localhost:6080/user", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != 401 {
		t.Errorf("Should have been rejected with 401, but got %v", response.Code)
	}
	actual := response.Body.String()
	if actual != expectedRejection{t.Errorf("Expected '%s', but got '%s'", expectedRejection, actual)}

	// Next, I should be able to log in
	request, _ = http.NewRequest(http.MethodPost, "http://localhost:6080/login", strings.NewReader(loginString))
	response = httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != 200 {t.Errorf("Should have been accepted with 200, but got %v", response.Code)}
	cookieValue := response.Header().Get("Set-Cookie")
	if !startsWith(cookieValue, "token="){
		t.Errorf("Should have got 'token=...', but got %v", cookieValue)
	}

	// Finally, make the original call, but with the cookieValue set
	request, _ = http.NewRequest(http.MethodGet, "http://localhost:6080/user/0", nil)
	request.Header.Set("Cookie", cookieValue)
	response = httptest.NewRecorder()

	server.ServeHTTP(response, request)
	if response.Code != 200 {
		t.Errorf("Should have been accepted with 200, but got %v", response.Code)
	}
	actual = response.Body.String()
	if actual != expectedSuccess {
		t.Errorf("Expected '%s', but got '%s'", expectedSuccess, actual)
	}
}

func TestInvalidHttpMethods(t *testing.T) {
	server := &LittleServer{}
	server.SetUpLogging(false, false)

	expected := `{"error":"http method not supported"}`

	notHandled := []string{
		http.MethodConnect,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodPut,
	}

	for _, method := range notHandled {
		t.Run(method, func(t *testing.T) {
			request, _ := http.NewRequest(method, "http://localhost:6080/picnic", nil)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)

			if response.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected %d, but got %d\r\n", http.StatusMethodNotAllowed, response.Code)
			}

			actual := response.Body.String()
			if actual != expected {
				t.Errorf("Expected\r\n    %v\r\nBut got\r\n    %v\r\n", expected, actual)
			}
		})
	}
}

func TestLittleServer_ServeHTTP(t *testing.T) {
	t.Run("Do a get", func(t *testing.T) {
		server := &LittleServer{}
		server.SetUpLogging(false, false)
		request, _ := http.NewRequest(http.MethodGet, "http://localhost:6080/api/people/123", nil)
		response := httptest.NewRecorder() // built-in mocks! :-D

		server.ServeHTTP(response, request)

		expected := `{"error":"page not found"}`
		actual := response.Body.String()

		if actual != expected {
			t.Errorf("Expected\r\n    %v\r\nBut got\r\n    %v\r\n", expected, actual)
		}
	})
}

func BenchmarkCallingServer(b *testing.B) {
	server := &LittleServer{}
	server.SetUpLogging(true, true) // if you don't limit to important, you'll get many MiB of logs
	request, _ := http.NewRequest(http.MethodGet, "http://localhost:6080/picnic", nil)
	response := httptest.NewRecorder()

	b.ResetTimer() // const stuff out of timed section. Not hugely necessary

	for i := 0; i < b.N; i++ {
		server.ServeHTTP(response, request)
	}
}

func startsWith(haystack, needle string)bool{return strings.Index(haystack, needle) == 0 }
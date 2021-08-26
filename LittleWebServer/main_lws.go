package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const(
	httpPort = ":6080"
)


type MyInputType struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Age int `json:"age"`
}

type LittleServer struct {

}

func main(){
	fmt.Printf("Bringing up a server on http://localhost%s\r\n", httpPort)

	server := &LittleServer{}
	http.Handle("/", server)

	err := http.ListenAndServe(httpPort, nil)
	if err != nil {
		fmt.Printf("Server failed: %v", err)
	}
}

func (serv *LittleServer)ServeHTTP(response http.ResponseWriter, request *http.Request){
	// should never modify `request`
	// `panic()` is restricted to the current request

	fmt.Printf("REQ/%s %s %s [%v]\r\n",request.Method, request.Host, request.RequestURI, request.Header)
	if request.Method == "POST" {
		http.MaxBytesReader(response, request.Body, 0xFFFF)
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields() // strict mode
		expectedStruct := MyInputType{}
		if err := decoder.Decode(&expectedStruct); err != nil {
			fmt.Printf("    Bad struct: %v\r\n", err)
			invalidInput(response)
			return
		} else {
			fmt.Printf("    Read struct: %v\r\n", expectedStruct)
		}
	}

	switch request.RequestURI {
	case "/panic":
		panic("Aaaaa!")
	case "/picnic":
		picnic(response)
	case "/favicon.ico":
		sendIcon(response)

	default:
		notFound(response)
	}
}

func picnic(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	_, err := response.Write([]byte(`{"message":"hello world"}`))
	if err != nil {panic(err)}
}

func notFound(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusNotFound)
	_, err := response.Write([]byte(`{"error":"page not found"}`))
	if err != nil {panic(err)}
}

func invalidInput(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusBadRequest)
	_, err := response.Write([]byte(`{"error":"input is invalid"}`))
	if err != nil {panic(err)}
}


func sendIcon(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "image/svg+xml")
	response.WriteHeader(http.StatusOK)
	_, err := response.Write([]byte(`<?xml version="1.0" standalone="no"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 480 150" height="64" width="64"><path d="m0 35.5l6.5-13 9.5 14.5 7-13 11.8 19.7 7.7-13.7 7.8 17 9.4-19.3 9.3 19.3 16-29.3 13.3 21.3 14.7-29.3 14.7 32.6 8.6-18.6 10.7 20.6 11.3-24 12 20 7.4-14.6 12 17.3 10-22 8 14 11.3-24 14 26 7.3-13.3 10.7 19.3 12-24.7 9.7 15 10.3-23.3 12 22.3 6.3-9.3 10.4 14 12-29.3 15.6 31.3 7-13.3 10 16.6 13.4-27.3 6.6 10.7 7.7-16.7 9 19.3 7.3-9.3 11.4 19.3 9.3-17.3 13.3 22 10.7-18 8 11.3 11.3-18 11.9 22 3.8-6.8v181.5h-480v-179.5z" fill="#175720"/></svg>`))
	if err != nil {panic(err)}
}
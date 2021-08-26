package main

import (
	"fmt"
	"net/http"
)

const(
	httpPort = ":6080"
)

type littleServer struct {

}

func main(){
	fmt.Printf("Bringing up a server on http://localhost%s\r\n", httpPort)

	server := &littleServer{}
	http.Handle("/", server)

	err := http.ListenAndServe(httpPort, nil)
	if err != nil {
		fmt.Printf("Server failed: %v", err)
	}
}

func (serv *littleServer)ServeHTTP(response http.ResponseWriter, request *http.Request){
	// should never modify `request`
	// `panic()` is restricted to the current request

	fmt.Printf("REQ/%s %s %s [%v]\r\n",request.Method, request.Host, request.RequestURI, request.Header)
	if request.Method == "POST"{
		fmt.Printf("    %v\r\n", request.Body)
	}

	response.WriteHeader(http.StatusOK)
	response.Header().Set("Content-Type", "application/json")
	response.Write([]byte(`{"message":"hello world"}`))
}
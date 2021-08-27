package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	userDb map[int]MyInputType
}

func main(){
	fmt.Printf("Bringing up a server on http://localhost%s\r\n", httpPort)

	server := &LittleServer{
		userDb: map[int]MyInputType{},
	}
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

	// URL property has query parser built in
	//request.URL.Query().Get("query-key") // => "query-value", given https://.../my/path?query-key=query-value

	pathBits := strings.Split(request.URL.Path, "/")
	for len(pathBits) > 0 && pathBits[0] == "" { // trim empty sections
		pathBits = pathBits[1:]
	}

	switch request.Method {
	case http.MethodGet:
		getHandler(serv, response, pathBits)
	case http.MethodPost:
		postHandler(serv, response, request, pathBits)
	default:
		unsupportedMethod(response)
	}
}

func postHandler(serv *LittleServer, response http.ResponseWriter, request *http.Request, pathBits []string) {
	if len(pathBits) < 1{
		invalidInput(response)
		return
	}

	switch pathBits[0] {
	case "user":
		postUser(serv, pathBits[1:], response, request)

	default:
		notFound(response)
	}
}

func getHandler(serv *LittleServer, response http.ResponseWriter, pathBits []string) {
	if len(pathBits) < 1{
		homePage(response)
		return
	}

	switch pathBits[0] {
	case "user":
		getUser(serv, pathBits[1:], response)

	case "panic":
		panic("panic!")
	case "picnic":
		picnic(response)
	case "favicon.ico":
		sendIcon(response)

	default:
		notFound(response)
	}
}

func postUser(serv *LittleServer, path []string, response http.ResponseWriter, request *http.Request) {
	if len(path) != 1 {invalidInput(response); return}
	id,err := strconv.Atoi(path[0])
	if err != nil {invalidInput(response); return}

	if request.Method == "POST" {
		http.MaxBytesReader(response, request.Body, 0xFFFF)
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields() // strict mode
		incomingUser := MyInputType{}
		if err = decoder.Decode(&incomingUser); err != nil {
			fmt.Printf("    Bad struct: %v\r\n", err)
			invalidInput(response)
			return
		} else {
			fmt.Printf("    Read struct: %v\r\n", incomingUser)
			serv.userDb[id] = incomingUser
		}
	}
}

func getUser(serv *LittleServer, path []string, response http.ResponseWriter) {
	if len(path) < 1 {
		listAllUsers(serv, response)
		return
	}

	// parse next part as int, return it
	if id,err:=strconv.Atoi(path[0]); err != nil{
		invalidInput(response)
		return
	}else{
		if data, err2 := json.Marshal(serv.userDb[id]); err2 != nil{
			panic(err2)
		} else {
			pWrite(data, response)
		}
	}
}

func listAllUsers(serv *LittleServer, response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)

	data, err := json.Marshal(serv.userDb)
	if err != nil {panic(err)}
	pWrite(data, response)
}

func pWrite(msg []byte, response http.ResponseWriter){
	if _, err := response.Write(msg); err != nil {panic(err)}
}

func unsupportedMethod(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusMethodNotAllowed)
	_, err := response.Write([]byte(`{"error":"http method not supported"}`))
	if err != nil {panic(err)}
}

func homePage(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	_, err := response.Write([]byte(`{"message":"hello world"}`))
	if err != nil {panic(err)}
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
	_, err := response.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><svg width="210mm" height="297mm" version="1.1" viewBox="0 0 210 297" xmlns="http://www.w3.org/2000/svg"><circle cx="26" cy="24" r="18" fill="#5b86bf"/></svg>`))
	if err != nil {panic(err)}
}
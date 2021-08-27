package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const(
	httpPort = ":6080"
)

//<editor-fold desc="Boiler plate">

var loginDetails = map[string]string{
	"ieb":"correct",
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"un"`
	jwt.StandardClaims
}

var jwtKey = []byte("Only suitable for development")

type MyInputType struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Age int `json:"age"`
}

type LittleServer struct {
	userDb     map[int]MyInputType
	lastSignIn time.Time
}

var infoLog *log.Logger
var warnLog *log.Logger
var critLog *log.Logger

func main(){
	logFile := SetUpLogging(false, false)
	defer func(file *os.File) { if file == nil {return}; _ = file.Close() }(logFile)

	infoLog.Printf("Bringing up a server on http://localhost%s\r\n", httpPort)

	server := &LittleServer{
		userDb: map[int]MyInputType{},
	}
	// add a sample user at [0]
	server.userDb[0] = MyInputType{
		ID:   -1,
		Name: "Sample user",
		Age:  20,
	}

	http.Handle("/", server)

	// just to test the JWT import is ok
	infoLog.Printf("JWT time: %v",jwt.TimeFunc())

	err := http.ListenAndServe(httpPort, nil)
	if err != nil {
		critLog.Fatalf("Server failed: %v", err)
	}
}

func SetUpLogging(useLogFile, onlyImportant bool) (usingFile *os.File) {
	logTarget := os.Stderr
	if useLogFile {
		file, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("could not open log file: %v", err)
		}
		logTarget = file
		usingFile = file
		if cwd, err := os.Getwd();err == nil {
			fmt.Printf("Logs going to %s\\%s\r\n", cwd, file.Name())
		} else {
			fmt.Printf("Logs going to %s\r\n", file.Name())
		}
		log.SetOutput(file)
	}

	infoLog = log.New(logTarget, "INFO:  ", log.Ldate|log.Ltime)
	warnLog = log.New(logTarget, "WARN:  ", log.Ldate|log.Ltime)
	critLog = log.New(logTarget, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	if onlyImportant {
		infoLog.SetOutput(ioutil.Discard)
	}

	return
}

func (serv *LittleServer)ServeHTTP(response http.ResponseWriter, request *http.Request){
	// should never modify `request`
	// `panic()` is restricted to the current request
	infoLog.Printf("REQ/%s %s %s [%v]\r\n", request.Method, request.Host, request.URL.Path, request.Header)

	// URL property has query parser built in
	//request.URL.Query().Get("query-key") // => "query-value", given https://.../my/path?query-key=query-value

	// Break path into bits, trim off empty sections at the start
	pathBits := strings.Split(request.URL.Path, "/")
	for len(pathBits) > 0 && pathBits[0] == "" { // trim empty sections
		pathBits = pathBits[1:]
	}

	switch request.Method {
	case http.MethodGet:
		getHandler(serv, response, request, pathBits)
	case http.MethodPost:
		postHandler(serv, response, request, pathBits)
	default:
		unsupportedMethod(request.Method, response)
	}
}

//</editor-fold>

func postHandler(serv *LittleServer, response http.ResponseWriter, request *http.Request, pathBits []string) {
	if len(pathBits) < 1{
		invalidInput(response)
		return
	}

	switch pathBits[0] {
	case "user":
		if !isAuthenticated(request, response) {return}
		postUser(serv, pathBits[1:], response, request)

	case "login":
		handleLogin(serv, response, request)

	default:
		notFound(response)
	}
}

func getHandler(serv *LittleServer, response http.ResponseWriter, request *http.Request, pathBits []string) {
	if len(pathBits) < 1{
		homePage(response)
		return
	}

	switch pathBits[0] {
	case "user":
		//if !isAuthenticated(request, response) {return}
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

func handleLogin(serv *LittleServer, response http.ResponseWriter, request *http.Request) {

	http.MaxBytesReader(response, request.Body, 0xFFFF)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields() // strict mode
	suppliedCreds := Credentials{}

	// Read input
	if err := decoder.Decode(&suppliedCreds); err != nil {
		warnLog.Printf("    Bad login struct: %v\r\n", err)
		invalidInput(response)
		return
	}

	// check password (very badly)
	if knownPassword, ok := loginDetails[suppliedCreds.Username]; !ok {
		invalidInput(response)
		return
	} else if knownPassword != suppliedCreds.Password {
		invalidInput(response)
		return
	}

	// Log-in is correct, build and return a token
	expiration := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: suppliedCreds.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiration.Unix(),
			Issuer:    "LittleWebServer",
			Subject:   suppliedCreds.Username,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(jwtKey)
	if err != nil {
		warnLog.Panicf("Could not create token: %v", err)
		return
	}

	http.SetCookie(response, &http.Cookie{
		Name:    "token",
		Value:   tokenStr,
		Expires: expiration,
	})

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	pWrite([]byte(`{"message":"ok"}`), response)
	serv.lastSignIn = time.Now()
}

func postUser(serv *LittleServer, path []string, response http.ResponseWriter, request *http.Request) {
	if len(path) != 1 {invalidInput(response); return}
	id,err := strconv.Atoi(path[0])
	if err != nil {invalidInput(response); return}

	http.MaxBytesReader(response, request.Body, 0xFFFF)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields() // strict mode
	incomingUser := MyInputType{}
	if err = decoder.Decode(&incomingUser); err != nil {
												   warnLog.Printf("    Bad struct: %v\r\n", err)
												   invalidInput(response)
												   return
												   } else {
			infoLog.Printf("    Read struct: %v\r\n", incomingUser)
			serv.userDb[id] = incomingUser
		}
}

func isAuthenticated(request *http.Request, response http.ResponseWriter) bool {
	tokenCookie,err := request.Cookie("token")
	if err != nil{
		if err == http.ErrNoCookie{
			mustAuth(response)
			return false
		}
	}

	claims := &Claims{}
	token,err := jwt.ParseWithClaims(tokenCookie.Value,
		claims,
		func(token *jwt.Token) (interface{}, error) {return jwtKey,nil },
	)
	if err != nil {
		if err == jwt.ErrSignatureInvalid{
			warnLog.Printf("JWT token does not match: %v", err)
		}
		infoLog.Printf("Failed to parse token: %v", err)
		mustAuth(response)
		return false
	}
	if !token.Valid {
		infoLog.Printf("User presented a signed but invalid token")
		mustAuth(response)
		return false
	}

	return true
}

func mustAuth(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusUnauthorized)
	pWrite([]byte(`{"error":"must provide token cookie"}`), response)
	infoLog.Printf("User supplied no auth token")
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
	} else {
		if userDetails,ok := serv.userDb[id]; !ok{
			notFound(response)
			return
		} else {
			if data, err2 := json.Marshal(userDetails); err2 != nil {
				warnLog.Panicf("Json marshal failed: %v", err2)
			} else {
				pWrite(data, response)
			}
		}
	}
}

func listAllUsers(serv *LittleServer, response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)

	data, err := json.Marshal(serv.userDb)
	if err != nil {warnLog.Panicf("Json marshal failed: %v",err)}
	pWrite(data, response)
}

func pWrite(msg []byte, response http.ResponseWriter){
	if _, err := response.Write(msg); err != nil {
		warnLog.Panicf("Failed to write response: %v", err)
	}
}

func unsupportedMethod(method string, response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusMethodNotAllowed)
	pWrite([]byte(`{"error":"http method not supported"}`), response)
	warnLog.Printf("Attempt to use %s which is not supported on this path", method)
}

func homePage(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	pWrite([]byte(`{"message":"hello world"}`), response)
}

func picnic(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	pWrite([]byte(`{"message":"hello world", "lunch":"tasty"}`), response)
}

func notFound(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusNotFound)
	pWrite([]byte(`{"error":"page not found"}`), response)
	warnLog.Printf("Attempted to access an invalid path")
}

func invalidInput(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusBadRequest)
	pWrite([]byte(`{"error":"input is invalid"}`), response)
	warnLog.Printf("User supplied malformed input")
}

func sendIcon(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "image/svg+xml")
	response.WriteHeader(http.StatusOK)
	_, err := response.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><svg version="1.1" viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg"><circle cx="24" cy="24" r="18" fill="#5b86bf"/></svg>`))
	if err != nil {warnLog.Panicf("Failed to write favicon: %v",err)}
}
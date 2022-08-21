package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

// ..

var router = mux.NewRouter()

func main() {

	router.HandleFunc("/", indexPageHandler)
	router.HandleFunc("/internal", internalPageHandler)
	router.HandleFunc("/login", loginHandler).Methods("POST")
	router.HandleFunc("/signup", signupHandler).Methods("POST")
	router.HandleFunc("/profile", profileHandler)
	router.HandleFunc("/profile", profileHandler).Methods("POST")
	router.HandleFunc("/logout", logoutHandler).Methods("POST")

	http.Handle("/", router)
	fmt.Println("Port Listening:8000")
	http.ListenAndServe(":8000", nil)
}

type form_method struct {
	LoginMethod   string
	LoginAction   string
	SignupMethod  string
	SignupAction  string
	ProfileAction string
	ProfileMethod string
}

func indexPageHandler(response http.ResponseWriter, request *http.Request) {
	parsedTemplate, _ := template.ParseFiles("public/index.html")
	form := form_method{LoginMethod: "POST", LoginAction: "/login", SignupMethod: "POST", SignupAction: "/signup"}
	err := parsedTemplate.Execute(response, form)
	if err != nil {
		log.Println("Error executing template :", err)
		return
	}
	// fmt.Fprintf(response, indexPage)
}

func profileHandler(response http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		parsedTemplate, _ := template.ParseFiles("public/profile.html")
		form := form_method{ProfileMethod: "POST", ProfileAction: "/profile"}
		err := parsedTemplate.Execute(response, form)
		if err != nil {
			fmt.Println("Error executing template :", err)
			log.Println("Error executing template :", err)
			return
		}
	} else {
		return
	}
}

const internalPage = `
<h1>Internal</h1>
<hr>
<small>User: %s</small>
<form method="post" action="/logout">
    <button type="submit">Logout</button>
</form>
`

func internalPageHandler(response http.ResponseWriter, request *http.Request) {
	userName := getUserName(request)
	if userName != "" {
		fmt.Fprintf(response, internalPage, userName)
	} else {
		http.Redirect(response, request, "/", 302)
	}
}

func signupHandler(response http.ResponseWriter, request *http.Request) {
	postBody, err := json.Marshal(map[string]string{
		"user_name": request.FormValue("user_name"),
		"email":     request.FormValue("email"),
		"password":  request.FormValue("password"),
	})
	if err != nil {
		fmt.Println("Can't serislize", request.Form)
	}
	URL := "http://127.0.0.0:8004/register"
	resp, err := handleRequest(request, URL, "POST", postBody)
	if err != nil {
		fmt.Println("--------err---------", err)
	}
	fmt.Println("--------resp---------", resp)
	redirectTarget := "/profile"
	http.Redirect(response, request, redirectTarget, 302)
}

func loginHandler(response http.ResponseWriter, request *http.Request) {
	name := request.FormValue("name")
	pass := request.FormValue("password")
	fmt.Println("-------request.FormValue--asd---------", request.Form)

	body, _ := json.Marshal(map[string]string{
		"user_name": "Test",
		"email":     "test@gmail.com",
		"password":  "password",
	})
	URL := "http://127.0.0.0:8004/register"
	resp, err := handleRequest(request, URL, "POST", body)
	fmt.Println("--------handleRequest---------------", resp, err)
	redirectTarget := "/"
	if name != "" && pass != "" {
		// .. check credentials ..
		setSession(name, response)
		redirectTarget = "/internal"
	}
	http.Redirect(response, request, redirectTarget, 302)
}

func logoutHandler(response http.ResponseWriter, request *http.Request) {
	clearSession(response)
	http.Redirect(response, request, "/", 302)
}

func setSession(userName string, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

func getUserName(request *http.Request) (userName string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

func handleRequest(request *http.Request, URL string, method string, data []byte) (response string, err error) {

	responseBody := bytes.NewBuffer(data)
	fmt.Println("-----------------data-------------------", string(data))
	client := &http.Client{}
	req, err := http.NewRequest(method, URL, responseBody)
	if err != nil {
		log.Fatalln(err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("test", "test")
	resp, err := client.Do(req)
	//Handle Error
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
		return "", err
	}
	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		return "", err
	}
	return string(body), nil

}

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

var (
	// Frontend Routers
	frontendDefaultRouter   = "/"
	frontendLoginRouter     = "/login"
	frontendInternalRouter  = "/internal"
	frontendSignupRouter    = "/signup"
	frontendProfileinRouter = "/profile"
	frontendLogoutRouter    = "/logout"

	// Backend Routers
	backendUrl            = "http://localhost:8004"
	basicAuthUsername     = "test"
	basicAuthPassowrd     = "test"
	backendLoginRouter    = fmt.Sprintf("%s/login", backendUrl)
	backendRegisterRouter = fmt.Sprintf("%s/register", backendUrl)
	backendProfileRouter  = fmt.Sprintf("%s/profile/", backendUrl)
)

func main() {

	var router = mux.NewRouter()
	router.HandleFunc(frontendDefaultRouter, indexPageHandler)
	router.HandleFunc(frontendInternalRouter, internalPageHandler)
	router.HandleFunc(frontendLoginRouter, loginHandler).Methods("POST")
	router.HandleFunc(frontendSignupRouter, signupHandler).Methods("GET", "POST")
	router.HandleFunc(frontendProfileinRouter, profileHandler).Methods("GET", "POST")
	router.HandleFunc(frontendLogoutRouter, logoutHandler).Methods("GET", "POST")

	http.Handle(frontendDefaultRouter, router)
	fmt.Println("Port Listening:8000")
	http.ListenAndServe(":8000", nil)
}

type UserData struct {
	Name      string
	Telephone string
	Email     string
}

type form_method struct {
	LoginMethod   string
	LoginAction   string
	SignupMethod  string
	SignupAction  string
	ProfileAction string
	ProfileMethod string
	FormError     string
	ProfileData   UserData
}

func indexPageHandler(response http.ResponseWriter, request *http.Request) {
	parsedTemplate, _ := template.ParseFiles("public/index.html")
	form := form_method{LoginMethod: "POST", LoginAction: frontendLoginRouter, SignupMethod: "POST", SignupAction: frontendSignupRouter}
	err := parsedTemplate.Execute(response, form)
	if err != nil {
		log.Println("Error executing template :", err)
		return
	}
}

func profileHandler(response http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		userName := getUserName(request)
		if userName == "" {
			http.Redirect(response, request, frontendDefaultRouter, http.StatusFound)
		}

		URL := backendProfileRouter + userName
		resp, err := handleRequest(response, request, URL, "GET", nil)
		if err != nil {
			handleError(response, err)
		}

		var jsonMap map[string]string
		err = json.Unmarshal([]byte(resp), &jsonMap)
		if err != nil {
			handleError(response, err)
		}
		parsedTemplate, _ := template.ParseFiles("public/profile.html")
		profile := UserData{Name: jsonMap["name"], Telephone: jsonMap["telephone"], Email: jsonMap["email"]}
		form := form_method{ProfileMethod: "POST", ProfileAction: frontendProfileinRouter, ProfileData: profile}
		err = parsedTemplate.Execute(response, form)
		if err != nil {
			log.Println("Error executing template :", err)
			return
		}
	} else {
		userName := getUserName(request)
		if userName != "" {
			http.Redirect(response, request, frontendDefaultRouter, http.StatusFound)
		}

		postBody, _ := json.Marshal(map[string]string{
			"name":      request.FormValue("name"),
			"telephone": request.FormValue("telephone"),
			"email":     request.FormValue("email"),
		})
		URL := backendProfileRouter + userName
		_, err := handleRequest(response, request, URL, "POST", postBody)
		if err != nil {
			handleError(response, err)
		}
		redirectTarget := "/profile"
		http.Redirect(response, request, redirectTarget, http.StatusFound)
	}
}

func handleError(response http.ResponseWriter, err error) {
	parsedTemplate, _ := template.ParseFiles("public/error_page.html")
	error_messege := form_method{FormError: fmt.Sprintf("%v", err)}
	fmt.Println("-------------error_messege--------------", error_messege)
	err = parsedTemplate.Execute(response, error_messege)
	if err != nil {
		log.Println("Error executing template :", err)
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
		http.Redirect(response, request, frontendDefaultRouter, http.StatusFound)
	}
}

func signupHandler(response http.ResponseWriter, request *http.Request) {
	postBody, err := json.Marshal(map[string]string{
		"user_name": request.FormValue("user_name"),
		"password":  request.FormValue("password"),
	})
	if err != nil {
		handleError(response, err)
	}

	_, err = handleRequest(response, request, backendRegisterRouter, "POST", postBody)
	if err != nil {
		handleError(response, err)
	}
	redirectTarget := frontendProfileinRouter
	http.Redirect(response, request, redirectTarget, http.StatusFound)
}

func loginHandler(response http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		username := request.FormValue("user_name")
		pass := request.FormValue("password")
		body, _ := json.Marshal(map[string]string{
			"user_name": username,
			"password":  pass,
		})
		_, err := handleRequest(response, request, backendLoginRouter, "POST", body)
		redirectTarget := frontendProfileinRouter
		if err != nil {
			handleError(response, err)
			redirectTarget = frontendInternalRouter
		}
		// .. check credentials ..
		setSession(username, response)
		http.Redirect(response, request, redirectTarget, http.StatusFound)
	}
}

func logoutHandler(response http.ResponseWriter, request *http.Request) {
	clearSession(response)
	http.Redirect(response, request, frontendDefaultRouter, http.StatusFound)
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

func handleRequest(response http.ResponseWriter, request *http.Request, URL string, method string, data []byte) (resp_data string, err error) {

	responseBody := bytes.NewBuffer(data)
	client := &http.Client{}
	req, err := http.NewRequest(method, URL, responseBody)
	if err != nil {
		log.Fatalln(err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(basicAuthUsername, basicAuthPassowrd)
	fmt.Println("Sending request to ", URL)
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

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
	router.HandleFunc("/signup", signupHandler).Methods("GET", "POST")
	router.HandleFunc("/profile", profileHandler).Methods("GET", "POST")
	router.HandleFunc("/logout", logoutHandler).Methods("GET", "POST")

	http.Handle("/", router)
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
	ProfileData   UserData
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
		userName := getUserName(request)
		fmt.Println("----------Profile-GET----------------------", userName)
		if userName == "" {
			http.Redirect(response, request, "/", http.StatusFound)
		}

		URL := fmt.Sprintf("http://127.0.0.0:8004/profile/%s", userName)
		resp, _ := handleRequest(request, URL, "GET", nil)
		fmt.Println("-----------------------", resp)

		var jsonMap map[string]string
		err := json.Unmarshal([]byte(resp), &jsonMap)
		if err != nil {
			fmt.Println("----------err-------------", jsonMap)
		}
		parsedTemplate, _ := template.ParseFiles("public/profile.html")
		profile := UserData{Name: jsonMap["name"], Telephone: jsonMap["telephone"], Email: jsonMap["email"]}
		form := form_method{ProfileMethod: "POST", ProfileAction: "/profile", ProfileData: profile}
		err = parsedTemplate.Execute(response, form)
		if err != nil {
			fmt.Println("Error executing template :", err)
			log.Println("Error executing template :", err)
			return
		}
	} else {
		fmt.Println("--------------nasdasd------------------", request.FormValue("name"))
		userName := getUserName(request)
		// if userName != "" {
		// 	fmt.Fprintf(response, internalPage, userName)
		// } else {
		// 	http.Redirect(response, request, "/", http.StatusFound)
		// }

		postBody, _ := json.Marshal(map[string]string{
			"name":      request.FormValue("name"),
			"telephone": request.FormValue("telephone"),
			"email":     request.FormValue("email"),
		})
		fmt.Println("--------------nasdasd------------------", string(postBody))
		URL := fmt.Sprintf("http://127.0.0.0:8004/profile/%s", userName)
		resp, err := handleRequest(request, URL, "POST", postBody)
		if err != nil {
			fmt.Println("--------err---------", resp, err)
		}
		redirectTarget := "/profile"
		http.Redirect(response, request, redirectTarget, http.StatusFound)
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
		http.Redirect(response, request, "/", http.StatusFound)
	}
}

func signupHandler(response http.ResponseWriter, request *http.Request) {
	postBody, err := json.Marshal(map[string]string{
		"user_name": request.FormValue("user_name"),
		"password":  request.FormValue("password"),
	})
	if err != nil {
		fmt.Println("Can't serislize", request.Form)
	}
	URL := "http://127.0.0.0:8004/register"
	resp, err := handleRequest(request, URL, "POST", postBody)
	fmt.Println("--------err---------", resp)
	if err != nil {
		fmt.Println("--------err---------", resp, err)
	}
	redirectTarget := "/profile"
	http.Redirect(response, request, redirectTarget, http.StatusFound)
}

func loginHandler(response http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		username := request.FormValue("user_name")
		pass := request.FormValue("password")
		fmt.Println("-------request.FormValue--asd---------", request.Form)

		body, _ := json.Marshal(map[string]string{
			"user_name": username,
			"password":  pass,
		})
		URL := "http://127.0.0.0:8004/login"
		resp, err := handleRequest(request, URL, "POST", body)
		fmt.Println("--------handleRequest---------------", resp, err)
		redirectTarget := "/profile"
		if err != nil {
			redirectTarget = "/internal"
		}
		// .. check credentials ..
		setSession(username, response)
		fmt.Println("------Usernam--------", getUserName(request))
		http.Redirect(response, request, redirectTarget, http.StatusFound)
	}
}

func logoutHandler(response http.ResponseWriter, request *http.Request) {
	clearSession(response)
	http.Redirect(response, request, "/", http.StatusFound)
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

package views

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

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

func RegisterRoutes() {
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

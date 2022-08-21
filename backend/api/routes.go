package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterRoutes() {

	r := mux.NewRouter()

	r.HandleFunc("/check", HealthCheck)
	r.Use(basicAuth)
	r.HandleFunc("/register", Register).Methods("POST")
	r.HandleFunc("/login", login).Methods("POST")
	r.HandleFunc("/profile/{username}", profile).Methods("GET", "POST")

	fmt.Println("Port Listening:8004")
	log.Fatal(http.ListenAndServe(":8004", r))
}

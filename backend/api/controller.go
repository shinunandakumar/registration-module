package api

import (
	"backend-reg/models"
	"backend-reg/mydb"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type JsonResponse struct {
	Status  int    `json:"status_code"`
	Message string `json:"message"`
}

var db = mydb.Conn()

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "API is up and running")
}

func Register(rw http.ResponseWriter, r *http.Request) {
	var user models.User

	response := JsonResponse{}
	var validate = validator.New()
	//validate the request body
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		response = JsonResponse{Status: http.StatusBadRequest, Message: "Invalid request"}
		json.NewEncoder(rw).Encode(response)
		return
	}

	//use the validator library to validate required fields
	// email and username validation #TODO
	if validationErr := validate.Struct(&user); validationErr != nil {
		rw.WriteHeader(http.StatusBadRequest)
		response = JsonResponse{Status: http.StatusBadRequest, Message: "Invalid request"}
		json.NewEncoder(rw).Encode(response)
		return
	}

	// Password hashing #TODO
	statement := fmt.Sprintf("INSERT INTO users (username,email,password) VALUES ('%s','%s','%s')", user.Username, user.Email, user.Password)
	db.QueryRow(statement)

	var err error
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		response = JsonResponse{Status: http.StatusInternalServerError, Message: "Internal Server error"}
		json.NewEncoder(rw).Encode(response)
		return
	}

	rw.WriteHeader(http.StatusCreated)
	response = JsonResponse{Status: http.StatusCreated, Message: "Success"}
	json.NewEncoder(rw).Encode(response)
}

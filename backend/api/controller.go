package api

import (
	"backend-reg/models"
	"backend-reg/mydb"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/go-playground/validator/v10"
)

type JsonResponse struct {
	Status  int    `json:"status_code"`
	Message string `json:"message"`
	Error   error  `json:"error_code"`
}

type Login struct {
	Username string `json:"user_name,omitempty" validate:"required"`
	Password string `json:"password,omitempty" validate:"required"`
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
		response = JsonResponse{Status: http.StatusBadRequest, Message: "Invalid request1", Error: err}
		json.NewEncoder(rw).Encode(response)
		return
	}

	//use the validator library to validate required fields
	// email and username validation #TODO
	if validationErr := validate.Struct(&user); validationErr != nil {
		rw.WriteHeader(http.StatusBadRequest)
		response = JsonResponse{Status: http.StatusBadRequest, Message: "Invalid request2", Error: validationErr}
		json.NewEncoder(rw).Encode(response)
		return
	}
	var err error
	passwordHash, _ := HashPassword(user.Password)
	statement := fmt.Sprintf("INSERT INTO users (username,password) VALUES ('%s','%s')", user.Username, passwordHash)
	_, err = db.Exec(statement)

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		response = JsonResponse{Status: http.StatusInternalServerError, Message: "Internal Server error", Error: err}
		json.NewEncoder(rw).Encode(response)
		return
	}

	rw.WriteHeader(http.StatusCreated)
	response = JsonResponse{Status: http.StatusCreated, Message: "Success", Error: err}
	json.NewEncoder(rw).Encode(response)
}

func login(rw http.ResponseWriter, r *http.Request) {
	login := Login{}
	response := JsonResponse{}
	// user := models.User
	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		response = JsonResponse{Status: http.StatusBadRequest, Message: "Invalid request1", Error: err}
		json.NewEncoder(rw).Encode(response)
		return
	}
	// Login Code here
	incoming_password := login.Password
	statement := fmt.Sprintf("SELECT username,password FROM users WHERE username='%s' LIMIT 1", login.Username)
	rows, err := db.Query(statement)
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&login.Username, &login.Password)
		if err != nil {
			log.Fatal(err)
		}
	}
	passwordmatch := CheckPasswordHash(incoming_password, login.Password)
	fmt.Println("passwordmatch", passwordmatch)
	if passwordmatch {
		rw.WriteHeader(http.StatusCreated)
		response = JsonResponse{Status: http.StatusCreated, Message: "Login Success", Error: err}
		json.NewEncoder(rw).Encode(response)

	} else {
		err = errors.New("password missmatch")
		rw.WriteHeader(http.StatusCreated)
		response = JsonResponse{Status: http.StatusBadRequest, Message: "Login Failed", Error: err}
		json.NewEncoder(rw).Encode(response)

	}

}

func profile(rw http.ResponseWriter, r *http.Request) {
	response := JsonResponse{}
	var err error
	username := mux.Vars(r)["username"]
	var userprofile models.UserProfile
	if r.Method == "GET" {
		// Get functions here
		statement := fmt.Sprintf("SELECT username,email,name,phone FROM users WHERE username='%s' LIMIT 1", username)
		err = db.QueryRow(statement).Scan(&userprofile.Username, &userprofile.Email, &userprofile.Name, &userprofile.Phone)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(userprofile)
		rw.WriteHeader(http.StatusOK)
		if (models.UserProfile{} == userprofile) {
			response = JsonResponse{Status: http.StatusBadRequest, Message: "No results Found", Error: err}
			json.NewEncoder(rw).Encode(response)
		} else {
			json.NewEncoder(rw).Encode(userprofile)
		}
		return
	} else {
		//validate the request body
		if err := json.NewDecoder(r.Body).Decode(&userprofile); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			response = JsonResponse{Status: http.StatusBadRequest, Message: "Invalid request1", Error: err}
			json.NewEncoder(rw).Encode(response)
			return
		}
		// Validation #TODO
		extra_update_q := ""
		if userprofile.Name != "" {
			extra_update_q += fmt.Sprintf(" name = '%s',", userprofile.Name)
		}
		if userprofile.Phone != "" {
			extra_update_q += fmt.Sprintf(" phone = '%s',", userprofile.Phone)
		}
		if userprofile.Email != "" {
			extra_update_q += fmt.Sprintf(" email = '%s' ", userprofile.Email)
		}
		if extra_update_q != "" {

			statement := "UPDATE users SET " + extra_update_q + fmt.Sprintf("WHERE username = '%s'", username)
			_, err = db.Exec(statement)
			rw.WriteHeader(http.StatusCreated)
			response = JsonResponse{Status: http.StatusCreated, Message: "Profile Updated", Error: err}
			json.NewEncoder(rw).Encode(response)
		} else {
			rw.WriteHeader(http.StatusForbidden)
			response = JsonResponse{Status: http.StatusForbidden, Message: "Ivalid inputs", Error: err}
			json.NewEncoder(rw).Encode(response)

		}
	}

}

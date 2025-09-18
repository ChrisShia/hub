package main

import (
	"authentication-service/internal/data"
	"net/http"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	//extract user from request
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var user data.User

	//Validate user

	//call UserModel.Insert
}

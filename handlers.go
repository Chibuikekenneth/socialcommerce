package main

import (
	"fmt"
	"time"
	"net/http"
)


func index_handler(response http.ResponseWriter, request *http.Request) {
	
	stmt, err := db.Prepare("INSERT INTO hits(route, ip, hdate) VALUES ($1, $2, $3)")
	fmt.Println("PREPARE ERROR: ", err)
	// TODO: record the errors into a logs table.

	result, err := stmt.Exec("index", request.RemoteAddr, time.Now())
	fmt.Println("RESULT: ", result)
	fmt.Println("ERROR: ", err)

	context.Message = ""
	// TODO: encapsulate the message variable so that it is a method instead of direct access.
	tpl.ExecuteTemplate(response, "index.html", context)	

}

func auth_handler(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	post_type := request.Form.Get("type")

	switch post_type {

	case "login":
		email := request.Form.Get("email")
		password := request.Form.Get("password")
		login(response, request, email, password)

	case "signup":
		email := request.Form.Get("email")
		password1 := request.Form.Get("password1")
		password2 := request.Form.Get("password2")
		signup(response, email, password1, password2)

	case "logout":
		clear_session(response, request)
		tpl.ExecuteTemplate(response, "index.html", nil)
	}
}

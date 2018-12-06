package main

import (
	"fmt"
	"time"
	"errors"
	"net/http"
	"math/rand"
	"net/smtp"
	"golang.org/x/crypto/bcrypt"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax = 63 / letterIdxBits
)

func RandString(n int) string {
	// META: Generates a random string.

	b := make([]byte, n)
	rand.Seed(time.Now().UTC().UnixNano())

	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func create_email_body(message string, route string, encrypted_code string, uid int) string {
	// META: Generates the string for email body.

	return fmt.Sprintf(`
			<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
			<html xmlns="http://www.w3.org/1999/xhtml">
			  <head>
			    <meta name="viewport" content="width=device-width"/>
			    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
				<style>
				.subscribe{
				  height: 70px;
				  text-align: center;
				}
				.button{
				  text-align: center;
				  font-size: 18px;
				  font-family: sans-serif;
				  font-weight: bold;
				  padding: 0 30px 0 30px;
				}
				.button a{
				  color: #FFFFFF;
				  text-decoration: none;
				}
				.buttonwrapper{
				  margin: 0 auto;
				}
				</style>
			</head>
			%v
			<tr class="subscribe">
			  <td style="padding: 20px 0 0 0;">
			    <table bgcolor="#009587" border="0" cellspacing="0" cellpadding="0" class="buttonwrapper">
			      <tr>
			        <td class="button" height="45">
			          <a href="http://localhost:8000/%v/%v/%v" target="_blank">Verify Your Account</a>
			        </td>
			      </tr>
			    </table>
			  </td>
			</tr>
	  </html>`, message, route, encrypted_code, uid)
}

func get_session(request *http.Request) (int, error) {
	// META: Gets the session cookie and returns the result.

	session, _ := store.Get(request, "session-name")
	untyped_uid, untyped_ok := session.Values["uid"]
	uid, ok := untyped_uid.(int)

	if !untyped_ok || !ok {
		return -1, errors.New("no session")
	} else {
		return uid, nil
	}
}

func set_session(response http.ResponseWriter, request *http.Request, uid int) {
	// META: Sets the session cookie and saves it.

	session, err := store.Get(request, "session-name")
	fmt.Println("SESSION GET ERROR:", err)
	session.Values["uid"] = uid
	err = session.Save(request, response)
	fmt.Println("SESSION SAVE ERROR:", err)
}

func clear_session(response http.ResponseWriter, request *http.Request) {
	// META: Clears the session so that it won't remember the user.


	session, _ := store.Get(request, "session-name")
	session.Values["uid"] = -1
	session.Save(request, response)
}

func execute_context(response http.ResponseWriter, templateName string, email string, eMsg string, hiddenMsg string) {
	//META: Creates Authentication context and returns the output into template.

	output := ClientContext{
		context,
		email,
		eMsg,
		hiddenMsg,
	}
	tpl.ExecuteTemplate(response, templateName, output)

}

func execute_engine(response http.ResponseWriter, html string, uid int) {
	// META: Creates Engine context and returns the output into template.

}

func login(response http.ResponseWriter, request *http.Request, email, password string) {

	if email == "" || password == "" {
		execute_context(response, "index.html", "", "Error: Please fill in all login information.", "login")
	}

	rows, _ := db.Query(fmt.Sprintf("SELECT uid, email, pw, created FROM users WHERE email = '%s'", email))
	defer rows.Close()

	if rows.Next() {

		var dbuid int
		var dbemail string
		var dbpassword string
		var dbcreated time.Time

		rows.Scan(&dbuid, &dbemail, &dbpassword, &dbcreated)
		err = bcrypt.CompareHashAndPassword([]byte(dbpassword), []byte(password))	

		if err != nil {
			execute_context(response, "index.html", "", "Error: Password is not correct.", "login")
		} else {
			set_session(response, request, dbuid)
			execute_engine(response, "engine.html", dbuid)
		}

	} else {
		execute_context(response, "index.html", "", "Error: Username does not exist.", "login")
	}
}

func signup(response http.ResponseWriter, email, password1, password2 string) {

	if password1 != password2 {
		execute_context(response, "index.html", email, "Error: Passwords do not match.", "signup")
	} else {
		query := fmt.Sprintf("SELECT uid, email FROM users WHERE verified = 'T' and email = '%s'", email)
		rows, _ := db.Query(query)

		if rows.Next() {
			var uid int
			var dbemail string
			rows.Scan(&uid, &email)
			var error_msg string = "Error: Email is already being used."
			execute_context(response, "index.html", dbemail, error_msg, "signup")

		} else {
			encrypted_password_bytes, _ := bcrypt.GenerateFromPassword([]byte(password1), 14)
			encrypted_password := string(encrypted_password_bytes)
			current_date := time.Now().Local().Format("2006-01-02")

	        var lastInsertId int
			err = db.QueryRow("INSERT INTO users (email, pw, created, verified, premium) VALUES ($1, $2, $3, $4, $5) returning uid;", email, encrypted_password, current_date, 'F', false).Scan(&lastInsertId)
			fmt.Println("ERROR INSERT: ", err)

			encrypted_code := RandString(20)
			stmt, err := db.Prepare(`INSERT INTO
								       	resets (uid, rtime, code, email, status)
								   	 VALUES((select distinct uid from users where email = $3 limit 1),$1,$2,$3,$4) returning rid;`)
			defer stmt.Close()
			fmt.Println(err)

			var rid int
			err = stmt.QueryRow(time.Now(), encrypted_code, email, true).Scan(&rid)
			fmt.Println("ERROR INSERT: ", err)
			fmt.Println("NEWLY CREATED RID: ", rid)

			subject := "Verify your account with Bizevote."

			body := create_email_body("To verify your account please press the button below.",
									  "verify",
									  encrypted_code,
									  rid)
			msg := "From: " + from + "\n" + 
				   "To: " + email + "\n" + 
				   "Subject: " + subject + "\n" +
				   "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
				   body
			smtp.SendMail("smtp.gmail.com:587",
						  smtp.PlainAuth("", from, "", "smtp.gmail.com"),
						  from, []string{email}, []byte(msg))
			execute_context(response, "index.html", email, "Success: an email has been sent to verify access and login.", "Success")
		}
	}
}

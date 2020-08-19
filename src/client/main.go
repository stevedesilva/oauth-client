package main

import (
	"fmt"
	"github.com/labstack/gommon/log"
	"html/template"
	"net/http"
)

type oauthType struct {
	authURL string
}

var oauth = oauthType{authURL: "http://www.google.com"}

func main() {
	fmt.Println("Oauth client")

	http.HandleFunc("/", home)
	http.HandleFunc("/login", login)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Error("Problem with server: ", err)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("template/index.html"))
	t.Execute(w, nil)

}

func login(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, oauth.authURL, http.StatusSeeOther)
}

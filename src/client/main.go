package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/labstack/gommon/log"
)

type oauthType struct {
	authURL   string
	logoutURL string
}

// from http://10.100.196.60:8080/auth/realms/silvade/.well-known/openid-configuration
var oauth = oauthType{
	authURL:   "http://10.100.196.60:8080/auth/realms/silvade/protocol/openid-connect/auth",
	logoutURL: "http://10.100.196.60:8080/auth/realms/silvade/protocol/openid-connect/logout",
}

type AppVar struct {
	AuthCode     string
	SessionState string
	State        string
}

var appVar = AppVar{}

var (
	t = template.Must(template.ParseFiles("src/client/template/index.html"))
)

func main() {
	fmt.Println("Oauth client")

	http.HandleFunc("/", home)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/loginAuthCodeCallback", loginAuthCodeCallback)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Error("Problem with server: ", err)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	//t := template.Must(template.ParseFiles("src/client/template/index.html"))
	err := t.Execute(w, nil)
	if err != nil {
		http.NotFound(w, r)
		return
	}

}

func login(w http.ResponseWriter, r *http.Request) {

	req, err := http.NewRequest("GET", oauth.authURL, nil)
	if err != nil {
		log.Print(err)
		return
	}

	q := url.Values{}
	q.Add("state", "123")
	q.Add("client_id", "billingApp")
	q.Add("response_type", "code")
	q.Add("redirect_uri", "http://localhost:8080/loginAuthCodeCallback")

	req.URL.RawQuery = q.Encode()
	http.Redirect(w, r, req.URL.String(), http.StatusSeeOther)
}

func loginAuthCodeCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	appVar.AuthCode = query.Get("code")
	appVar.SessionState = query.Get("session_state")
	appVar.State = query.Get("state")
	r.URL.RawQuery = ""
	fmt.Printf("Request queries : %+v\n", appVar)
	// use 303 instead of 302
	// https://en.wikipedia.org/wiki/List_of_HTTP_status_codes#3xx_Redirection
	http.Redirect(w, r, "http://localhost:8080/", http.StatusSeeOther)
}

func logout(w http.ResponseWriter, r *http.Request) {
	q := url.Values{}
	q.Add("redirect_uri", "http://localhost:8080")

	loURL, err := url.Parse(oauth.logoutURL)
	if err != nil {
		log.Print(err)
	}
	loURL.RawQuery = q.Encode()

	http.Redirect(w, r, loURL.String(), http.StatusSeeOther)

}

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"

	"github.com/stevedesilva/oauth-client/src/client/model"
)

type oauthConfig struct {
	appId                 string
	appPassword           string
	authURL               string
	logoutURL             string
	afterLogoutRedirect   string
	loginAuthCodeCallback string
	tokenEndpoint         string
}

// from http://10.100.196.60:8080/auth/realms/silvade/.well-known/openid-configuration
var config = oauthConfig{
	authURL:               "http://10.100.196.60:8080/auth/realms/silvade/protocol/openid-connect/auth",
	logoutURL:             "http://10.100.196.60:8080/auth/realms/silvade/protocol/openid-connect/logout",
	tokenEndpoint:         "http://10.100.196.60:8080/auth/realms/silvade/protocol/openid-connect/token",
	afterLogoutRedirect:   "http://localhost:8080/",
	loginAuthCodeCallback: "http://localhost:8080/loginAuthCodeCallback",
	appId:                 "billingApp",
	appPassword:           "c30285ef-395b-44be-8696-54f0cdc72582",
}

type AppVar struct {
	AuthCode       string
	SessionState   string
	State          string
	AccessToken    string
	RefreshToken   string
	Scope          string
	SessionStateEx string
}

var appVar = AppVar{}

var (
	t = template.Must(template.ParseFiles("src/client/template/index.html"))
)

func init() {
	log.SetFlags(log.Ltime)
}
func main() {
	fmt.Println("Oauth client")

	http.HandleFunc("/", addLog(home))
	http.HandleFunc("/login", addLog(login))
	http.HandleFunc("/logout", addLog(logout))
	http.HandleFunc("/loginAuthCodeCallback", addLog(loginAuthCodeCallback))
	http.HandleFunc("/exchangeToken", addLog(exchangeToken))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Problem with server: ", err)
	}
}

func addLog(fn func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handlerName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
		log.SetPrefix(fmt.Sprintf("%s\t", handlerName))
		log.Printf("--> %s  \n", handlerName)
		fn(w, r)
		log.Printf("<-- %s  \n", handlerName)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	//t := template.Must(template.ParseFiles("src/client/template/index.html"))
	err := t.Execute(w, appVar)
	if err != nil {
		log.Println(err)
		return
	}

}

func login(w http.ResponseWriter, r *http.Request) {

	req, err := http.NewRequest("GET", config.authURL, nil)
	if err != nil {
		log.Println(err)
		return
	}

	q := url.Values{}
	q.Add("state", "123")
	q.Add("client_id", config.appId)
	q.Add("response_type", "code")
	q.Add("redirect_uri", config.loginAuthCodeCallback)

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
	q.Add("redirect_uri", config.afterLogoutRedirect)

	loURL, err := url.Parse(config.logoutURL)
	if err != nil {
		log.Println(err)
	}
	loURL.RawQuery = q.Encode()
	appVar = AppVar{}
	http.Redirect(w, r, loURL.String(), http.StatusSeeOther)

}

/*


4.1.3.  Access Token Request

   The client makes a request to the token endpoint by sending the
   following parameters using the "application/x-www-form-urlencoded"
   format per Appendix B with a character encoding of UTF-8 in the HTTP
   request entity-body:

   grant_type
         REQUIRED.  Value MUST be set to "authorization_code".

   code
         REQUIRED.  The authorization code received from the
         authorization server.

   redirect_uri
         REQUIRED, if the "redirect_uri" parameter was included in the
         authorization request as described in Section 4.1.1, and their
         values MUST be identical.

   client_id
         REQUIRED, if the client is not authenticating with the
         authorization server as described in Section 3.2.1.

   If the client type is confidential or the client was issued client
   credentials (or assigned other authentication requirements), the
   client MUST authenticate with the authorization server as described
   in Section 3.2.1.
*/

func exchangeToken(w http.ResponseWriter, r *http.Request) {
	// Request
	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("code", appVar.AuthCode)
	form.Add("redirect_uri", config.loginAuthCodeCallback)
	form.Add("client_id", config.appId)
	req, err := http.NewRequest("POST", config.tokenEndpoint, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		log.Println(err)
		return
	}

	req.SetBasicAuth(config.appId, config.appPassword)
	//Client
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		log.Print("couldn't get access token ", err, "\n")
		return
	}

	// Process response
	byteBody, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Println(err)
		return
	}

	//model.AccessTokenResponse{}
	accessTokenResponse := &model.AccessTokenResponse{}
	json.Unmarshal(byteBody, accessTokenResponse)

	appVar.AccessToken = accessTokenResponse.AccessToken
	appVar.RefreshToken = accessTokenResponse.RefreshToken
	appVar.Scope = accessTokenResponse.Scope
	appVar.SessionStateEx = accessTokenResponse.SessionState

	log.Print(string(byteBody), "\n")
	err = t.Execute(w, appVar)
	if err != nil {
		log.Println(err)
		return
	}
}

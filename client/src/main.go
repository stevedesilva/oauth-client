package main

import (
	"fmt"
	"github.com/labstack/gommon/log"
	"net/http"
)

func main() {
	fmt.Println("Oauth client")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w,"Hello Oauth client!")
	})
	err := http.ListenAndServe(":8080",nil)
	if err != nil {
		log.Error(err)
	}
}

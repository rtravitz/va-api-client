package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/mux"
)

const callback = "https://go-va-api-client.herokuapp.com/auth/callback"

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Incoming request:\n%+v\n", r)
	w.Write([]byte("redirect route"))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse("https://dev-api.va.gov/oauth2/authorization")
	if err != nil {
		log.Println("Error parsing URL", err)
		return
	}

	q := u.Query()
	q.Add("client_id", os.Getenv("VA_OAUTH_CLIENT_ID"))
	q.Add("redirect_uri", callback)
	q.Add("response_type", "code")

	u.RawQuery = q.Encode()
	fmt.Println(u.String())

	http.Redirect(w, r, u.String(), http.StatusFound)
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	r := mux.NewRouter()

	s := r.PathPrefix("/auth").Subrouter()
	s.HandleFunc("/login", loginHandler)
	s.HandleFunc("/callback", redirectHandler)

	r.PathPrefix("/").Handler(fs)

	port := os.Getenv("PORT")
	if os.Getenv("ENV") == "LOCAL" {
		port = "localhost:" + port
	} else {
		port = ":" + port
	}

	log.Printf("Listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}

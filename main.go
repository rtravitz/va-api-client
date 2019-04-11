package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

var (
	clientID     = os.Getenv("VA_OAUTH_CLIENT_ID")
	clientSecret = os.Getenv("VA_OAUTH_CLIENT_SECRET")
	redirectURL  = "https://go-va-api-client.herokuapp.com/auth/callback"
)

func main() {
	ctx := context.Background()

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://dev-api.va.gov/oauth2/authorization",
			TokenURL: "https://dev-api.va.gov/oauth2/token",
		},
		RedirectURL: redirectURL,
		Scopes:      []string{"openid", "profile", "email"},
	}

	state := "foobar"

	r := mux.NewRouter()

	s := r.PathPrefix("/auth").Subrouter()
	s.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		authURL := config.AuthCodeURL(state)
		fmt.Println("AUTH URL", authURL)
		http.Redirect(w, r, config.AuthCodeURL(state), http.StatusFound)
	})

	s.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "state did not match", http.StatusBadRequest)
			return
		}

		oauth2Token, err := config.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		oauth2Token.AccessToken = "*REDACTED*"

		resp := struct {
			OAuth2Token   *oauth2.Token
			IDTokenClaims *json.RawMessage // ID Token payload is just JSON.
		}{oauth2Token, new(json.RawMessage)}

		data, err := json.MarshalIndent(resp, "", "    ")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(data)
	})

	fs := http.FileServer(http.Dir("static"))
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

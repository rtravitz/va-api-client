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
	clientID     = os.Getenv("VA_CLIENT")
	clientSecret = os.Getenv("VA_SECRET")
	baseURL      = os.Getenv("VA_API")
	redirectURL  = fmt.Sprintf("http://localhost:%s/auth/callback", os.Getenv("PORT"))
	accessToken  = ""
)

//OauthEndpoints are used as part of the oauth 2 flow and can be discovered
type OauthEndpoints struct {
	Auth  string `json:"authorization_endpoint"`
	Token string `json:"token_endpoint"`
}

func configureOauth() oauth2.Config {
	// Discovering endpoints to use in the oauth2 Authorization flow
	res, err := http.Get(baseURL + "/oauth2/.well-known/openid-configuration")
	if err != nil {
		log.Fatal("Failed to get oauth2 configuration:", err)
	}
	defer res.Body.Close()

	var endpoints OauthEndpoints
	err = json.NewDecoder(res.Body).Decode(&endpoints)
	if err != nil || endpoints.Auth == "" || endpoints.Token == "" {
		msg := "Failed to parse oauth2 endpoints.Err: %v, Auth: %s, Token: %s\n"
		log.Fatalf(msg, err, endpoints.Auth, endpoints.Token)
	}

	return oauth2.Config{
		// Client ID issued by the VA API Platform team
		ClientID: clientID,

		// Client Secret issued by the VA API Platform team
		ClientSecret: clientSecret,

		// Discovered endpoints
		Endpoint: oauth2.Endpoint{
			AuthURL:  endpoints.Auth,
			TokenURL: endpoints.Token,
		},

		// Redirect URL for the client. This must match the redirect URL provided at API signup
		RedirectURL: redirectURL,

		// Scopes must include openid
		Scopes: []string{"openid", "profile", "email", "service_history.read"},
	}
}

func callbackHandler(ctx context.Context, config oauth2.Config, state string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get state and make sure it matches the original
		if r.URL.Query().Get("state") != state {
			http.Error(w, "state did not match", http.StatusBadRequest)
			return
		}

		// Exchange code return in query param for an access token
		oauth2Token, err := config.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		accessToken = oauth2Token.AccessToken

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func serviceHistoryHandler(w http.ResponseWriter, r *http.Request) {
	serviceHistory, err := getServiceHistory(accessToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get or parse service history.")
		return
	}

	respondWithJSON(w, http.StatusOK, serviceHistory)
}

func loginHandler(config oauth2.Config, state string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, config.AuthCodeURL(state), http.StatusFound)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func main() {
	ctx := context.Background()
	config := configureOauth()

	// This state is just an example. A dynamic state should be used for each request.
	state := "foobar"

	r := mux.NewRouter()
	r.HandleFunc("/api/servicehistory", serviceHistoryHandler)
	r.HandleFunc("/auth/login", loginHandler(config, state))
	r.HandleFunc("/auth/callback", callbackHandler(ctx, config, state))

	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/").Handler(fs)

	port := os.Getenv("PORT")

	// Prevents OSX from asking if you want to accept incoming connections with every new binary
	if os.Getenv("ENV") == "LOCAL" {
		port = "localhost:" + port
	} else {
		port = ":" + port
	}

	log.Printf("Listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}

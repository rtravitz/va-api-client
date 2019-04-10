package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("redirect route"))
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	r := mux.NewRouter()
	r.HandleFunc("/redirect", redirectHandler)
	r.PathPrefix("/").Handler(fs)

	port := ":5000"
	log.Printf("Listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}

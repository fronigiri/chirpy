package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	mux.Handle("/assets/", http.FileServer(http.Dir(".")))
	log.Fatal(server.ListenAndServe())
}

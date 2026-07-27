package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))
	srv := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	fmt.Printf("starting server on http://localhost%v\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

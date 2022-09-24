package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello From myapp Service!\n"))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", HandleRoot)

	log.Fatal(http.ListenAndServe(":80", r))
}

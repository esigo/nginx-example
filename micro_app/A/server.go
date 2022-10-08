package main

import (
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello From myapp Service!\n"))
}

func HandleB(w http.ResponseWriter, r *http.Request) {
	url := "http://microapp-service-b.myapp.svc.cluster.local:80/test"
	http.NewRequest("GET", url, nil)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
	}
	req.Header = r.Header
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
		}
		bodyString := string(bodyBytes)
		log.Println(bodyString)
		w.Write(bodyBytes)
	} else {
		w.Write([]byte("Gorilla!\n"))
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", HandleRoot)

	log.Fatal(http.ListenAndServe(":80", r))
}

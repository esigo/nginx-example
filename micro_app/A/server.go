package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello From myapp Service!\n"))
}

type Payload struct {
	S string `json:"S"`
}

func HandleHello(w http.ResponseWriter, r *http.Request) {
	name := ""
	paths := strings.Split(r.URL.Path, "/")
	if len(paths) > 2 {
		name = paths[2]
	}
	data := Payload{
		S: name,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		w.Write([]byte("can't read the name!\n"))
		return
	}
	body := bytes.NewReader(payloadBytes)
	url := "http://microapp-service-b.myapp.svc.cluster.local:80/hello"

	req, err := http.NewRequest("GET", url, body)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	req.Header = r.Header
	req.Header.Set("Content-Type", "application/json")

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
	r.HandleFunc("/hello/{name:[a-z]+}", HandleHello)

	log.Fatal(http.ListenAndServe(":80", r))
}

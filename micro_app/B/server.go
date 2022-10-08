package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

type HelloService interface {
	Hello(string) (string, error)
}

type helloService struct{}

func (helloService) Hello(s string) (string, error) {
	if s == "" {
		return "", ErrEmpty
	}
	var sb strings.Builder

	sb.WriteString("hello ")
	sb.WriteString(s)
	sb.WriteString("!")
	return sb.String(), nil
}

var ErrEmpty = errors.New("Empty string")

type helloRequest struct {
	S string `json:"s"`
}

type helloResponse struct {
	V   string `json:"v"`
	Err string `json:"err,omitempty"`
}

func makeHelloEndpoint(svc HelloService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(helloRequest)
		v, err := svc.Hello(req.S)
		if err != nil {
			return helloResponse{v, err.Error()}, nil
		}
		return helloResponse{v, ""}, nil
	}
}

func decodeHelloRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request helloRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func main() {
	svc := helloService{}

	helloHandler := httptransport.NewServer(
		makeHelloEndpoint(svc),
		decodeHelloRequest,
		encodeResponse,
	)

	http.Handle("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":80", nil))
}

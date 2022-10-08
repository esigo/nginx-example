package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
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
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(helloRequest)
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(
			attribute.String("attr", "serve-b"),
		)
		defer span.End()
		v, err := svc.Hello(req.S)
		if err != nil {
			return helloResponse{v, err.Error()}, nil
		}
		return helloResponse{v, ""}, nil
	}
}

func decodeHelloRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	propagator := otel.GetTextMapPropagator()
	ctx = propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	_, span := otel.Tracer("service-b/hello").Start(ctx, "service-b hello")
	span.SetAttributes(
		attribute.String("attr", "serve-b"),
	)
	defer span.End()
	var request helloRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func initProvider() (func(context.Context) error, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("microapp-service-b"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, "otel-coll-collector.otel.svc:4317", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tracerProvider.Shutdown, nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	shutdown, err := initProvider()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider: %w", err)
		}
	}()
	svc := helloService{}

	helloHandler := httptransport.NewServer(
		makeHelloEndpoint(svc),
		decodeHelloRequest,
		encodeResponse,
	)

	http.Handle("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":80", nil))
}

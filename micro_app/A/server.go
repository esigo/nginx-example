package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello From myapp Service!\n"))
}

type Payload struct {
	S string `json:"S"`
}

func HandleHello(w http.ResponseWriter, r *http.Request) {
	propagator := otel.GetTextMapPropagator()
	ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	ctx, span := otel.Tracer("service-a/hello").Start(ctx, "service-a hello")
	span.SetAttributes(
		attribute.String("attr", "serve-b"),
	)
	defer span.End()

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
	r = r.WithContext(ctx)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))
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

func initProvider() (func(context.Context) error, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("microapp-service"),
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

	r := mux.NewRouter()
	r.HandleFunc("/", HandleRoot)
	r.HandleFunc("/hello/{name:[a-z]+}", HandleHello)

	log.Fatal(http.ListenAndServe(":80", r))
}

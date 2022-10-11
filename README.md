[![Image CI](https://github.com/esigo/nginx-example/actions/workflows/ci.yaml/badge.svg?branch=main&event=push)](https://github.com/esigo/nginx-example/actions/workflows/ci.yaml)

# nginx-example

This repo hosts a simple app to demonstrate distributed tracing feature of nginx.

```mermaid
graph TB
    subgraph Browser
    start["http://esigo.dev/hello/nginx"]
    end

    subgraph app
        sa[service-a]
        sb[service-b]
        sa --> |name: nginx| sb
        sb --> |hello nginx!| sa
    end

    subgraph otel
        otc["Otel Collector"] 
    end

    subgraph observability
        tempo["Tempo"]
        grafana["Grafana"]
        backend["Jaeger"]
    end

    subgraph ingress-nginx
        ngx[nginx]
    end

    subgraph ngx[nginx]
        ng[nginx]
        om[OpenTelemetry module]
    end

    subgraph Node
        app
        otel
        observability
        ingress-nginx
        om --> |otlp-gRPC| otc --> |jaeger| backend
        otc --> |otlp-gRPC| tempo --> grafana
        sa --> |otlp-gRPC| otc
        sb --> |otlp-gRPC| otc
        start --> ng --> sa
    end
```
build images:
```console
make images
```

deploy demo app:
```console
make deploy-app
```

deploy otel collector, grafan, tempo and Jaeger all-in-one:
```console
make helm-repo
make observability
```

test:
```console
kubectl port-forward --namespace=ingress-nginx service/ingress-nginx-controller 8090:80
bash test.sh
```

##  controller-configmap.yaml:

```yaml
otlp-collector-host: "otel-coll-collector.otel.svc"
```

or

```yaml
otlp-collector-host: "tempo.observability.svc"
```

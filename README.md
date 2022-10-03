[![Image CI](https://github.com/esigo/nginx-example/actions/workflows/ci.yaml/badge.svg?branch=main&event=push)](https://github.com/esigo/nginx-example/actions/workflows/ci.yaml)

# nginx-example

deploy demo app:

```console
cd micro_app
bash deploy-demo.sh
```

deploy otel collector, grafan, tempo and Jaeger all-in-one:
```console
cd observability/
bash backend.sh
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

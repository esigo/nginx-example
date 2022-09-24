#!/bin/bash
set -e
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.yaml
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
helm install otel-collector-operator --namespace otel --create-namespace open-telemetry/opentelemetry-operator
kubectl apply -f collector.yaml

kubectl create namespace observability
kubectl apply -f https://github.com/jaegertracing/jaeger-operator/releases/download/v1.37.0/jaeger-operator.yaml -n observability
kubectl apply -f jaeger.yaml --namespace observability

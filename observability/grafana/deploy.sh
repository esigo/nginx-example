#!/bin/bash
set -e

helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

helm upgrade --install tempo grafana/tempo --create-namespace --namespace observability
helm upgrade -f grafana-values.yaml --install grafana grafana/grafana --create-namespace --namespace observability

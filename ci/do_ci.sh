#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

kubectl cluster-info --context kind-chart-testing

kind load docker-image docker.io/library/service-a:last --name chart-testing
kind load docker-image docker.io/library/service-b:last --name chart-testing

docker images

make observability

clusterIp=$(docker container inspect chart-testing-control-plane \
--format '{{ .NetworkSettings.Networks.kind.IPAddress }}')
nodePort=$(kubectl get svc -n ingress-nginx ingress-nginx-controller -o jsonpath='{.spec.ports[?(@.name=="http")].nodePort}')

kubectl get pods -n ingress-nginx
kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx -n ingress-nginx

make deploy-app

echo " ${clusterIp} esigo.dev" | sudo tee -a /etc/hosts

kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx -n ingress-nginx
kubectl get services -n ingress-nginx

host=esigo.dev
port=${nodePort}

response=$(curl -s -o /dev/null -w "%{http_code}" http://${host}:${port}/hello/nginx)
kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx -n ingress-nginx

if [ ${response} -eq 200 ]; then
    echo "Ingress test passed. HTTP response code: ${response}"
else
    kubectl get events --all-namespaces --field-selector type=Warning,reason!=Successful
    kubectl get pods -n myapp
    kubectl get pods -n ingress-nginx
    curl http://${host}:${port}/hello/nginx
    kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx -n ingress-nginx
    sleep 10
    # exit 1
fi

curl http://${host}:${port}
curl http://${host}:${port}/hello/nginx
curl http://${host}:${port}/hello/nginx
curl http://${host}:${port}/hello/nginx

kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx -n ingress-nginx

kubectl get services -n observability
kubectl get pods -n observability

kubectl port-forward --namespace=observability service/jaeger-all-in-one-query 16686:16686&
sleep 40

curl -X GET http://localhost:16686/api/services

getTrace() {
    local serviceName=$1
    echo "Getting trace for ${serviceName}"
    local traceId=$(curl "http://localhost:16686/api/traces?service=${serviceName}&limit=1" | jq -r '.data[0].traceID')
    echo "trace Id: ${traceId}"
    curl http://localhost:16686/api/traces/${traceId}
    curl http://localhost:16686/api/traces/${traceId}?prettyPrint=true
}

getTrace "microapp-service"
getTrace "microapp-service-b"
getTrace "nginx-proxy"

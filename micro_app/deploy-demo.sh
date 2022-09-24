#!/bin/bash
set -e

docker build -f Dockerfile.app -t service-app:last .

kubectl create namespace myapp
kubectl apply -f myapp.yaml
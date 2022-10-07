.PHONY: images
images:
	- docker build -f micro_app/Dockerfile.app -t service-app:last micro_app

.PHONY: deploy-app
deploy-app:
	- kubectl apply -f micro_app/namespace.yaml
	- kubectl apply -f micro_app/myapp.yaml

.PHONY: helm-repo
helm-repo:
	- helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
	- helm repo add grafana https://grafana.github.io/helm-charts
	- helm repo update

.PHONY: observability
observability:
	- kubectl apply -f observability/namespace.yaml

	- echo "deploying otel collector"
	- kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.yaml
	- helm upgrade --install otel-collector-operator -n otel --create-namespace open-telemetry/opentelemetry-operator
	- kubectl apply -f observability/collector.yaml

	- echo "deploying Jaeger"
	- kubectl apply -f https://github.com/jaegertracing/jaeger-operator/releases/download/v1.37.0/jaeger-operator.yaml -n observability
	- kubectl apply -f observability/jaeger.yaml -n observability

	- echo "deploying grafana and tempo"
	- helm upgrade --install tempo grafana/tempo --create-namespace -n observability
	- helm upgrade -f observability/grafana/grafana-values.yaml --install grafana grafana/grafana --create-namespace -n observability

.PHONY: clean
clean:
	- kubectl delete -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.yaml
	- helm uninstall otel-collector-operator -n otel
	- kubectl delete -f observability/collector.yaml

	- kubectl delete -f https://github.com/jaegertracing/jaeger-operator/releases/download/v1.37.0/jaeger-operator.yaml -n observability
	- kubectl delete -f observability/jaeger.yaml -n observability

	- helm uninstall tempo grafana/tempo -n observability
	- helm uninstall grafana -n observability

	- kubectl delete namespace observability
	- kubectl delete namespace otel

	- kubectl delete -f micro_app/myapp.yaml
	- kubectl delete namespace myapp

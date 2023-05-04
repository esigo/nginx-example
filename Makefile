.PHONY: images
images:
	- docker build -f micro_app/A/Dockerfile.app -t service-a:last micro_app/A
	- docker build -f micro_app/B/Dockerfile.app -t service-b:last micro_app/B
	- docker build -f micro_app/C/Dockerfile.app -t service-c:last micro_app/C

.PHONY: deploy-app
deploy-app:
	- kubectl apply -f micro_app/namespace.yaml
	- kubectl apply -f micro_app/microapp-a.yaml
	- kubectl apply -f micro_app/microapp-b.yaml
	- kubectl apply -f micro_app/microapp-c.yaml

.PHONY: helm-repo
helm-repo:
	- helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
	- helm repo add grafana https://grafana.github.io/helm-charts
	- helm repo update

observability: helm-repo otel-collector jaeger zipkin grafana

.PHONY: otel-collector
otel-collector:
	echo "deploying otel collector"
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.yaml
	./wait.sh _wait "pod -l app=cert-manager -n cert-manager"
	./wait.sh _wait "pod -l app=cert-manager -n cert-manager"
	./wait.sh _wait "pod -l app=cainjector -n cert-manager"
	./wait.sh _wait "pod -l app=webhook -n cert-manager"

	helm upgrade --install otel-collector-operator -n otel --create-namespace open-telemetry/opentelemetry-operator
	./wait.sh _wait "pod -l app.kubernetes.io/name=opentelemetry-operator -n otel"

	kubectl apply -f observability/collector.yaml
	./wait.sh _wait "pod -l app.kubernetes.io/instance=otel.otel-coll -n otel"

.PHONY: jaeger
jaeger:
	kubectl apply -f observability/namespace.yaml
	echo "deploying Jaeger"
	kubectl apply -f https://github.com/jaegertracing/jaeger-operator/releases/download/v1.37.0/jaeger-operator.yaml -n observability
	./wait.sh _wait "pod -l name=jaeger-operator -n observability"
	kubectl apply -f observability/jaeger.yaml -n observability
	./wait.sh _wait "pod -l app=jaeger -n observability"

.PHONY: zipkin
zipkin:
	kubectl apply -f observability/namespace.yaml
	kubectl apply -f observability/zipkin.yaml -n observability
	./wait.sh _wait "pod -l app=zipkin -n observability"

.PHONY: grafana
grafana:
	kubectl apply -f observability/namespace.yaml
	echo "deploying grafana and tempo"
	helm upgrade --install tempo grafana/tempo --create-namespace -n observability
	./wait.sh _wait "pod -l app.kubernetes.io/name=tempo -n observability"
	helm upgrade -f observability/grafana/grafana-values.yaml --install grafana grafana/grafana --create-namespace -n observability
	./wait.sh _wait "pod -l app.kubernetes.io/name=grafana -n observability"

clean: clean-app clean-observability

.PHONY: clean-observability
clean-observability:
	- kubectl delete -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.yaml
	- helm uninstall otel-collector-operator -n otel
	- kubectl delete -f observability/collector.yaml

	- kubectl delete -f https://github.com/jaegertracing/jaeger-operator/releases/download/v1.37.0/jaeger-operator.yaml -n observability
	- kubectl delete -f observability/jaeger.yaml -n observability

	- kubectl delete -f observability/zipkin.yaml -n observability

	- helm uninstall tempo grafana/tempo -n observability
	- helm uninstall grafana -n observability

	- kubectl delete namespace observability
	- kubectl delete namespace otel

.PHONY: clean-app
clean-app:
	- kubectl delete -f micro_app/microapp-a.yaml
	- kubectl delete -f micro_app/microapp-b.yaml
	- kubectl delete namespace myapp

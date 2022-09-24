# nginx-example

deploy demo app:

```console
cd micro_app
bash deploy-demo.sh
```

deploy collector and Jaeger all-in-one:
```console
cd observability/
bash backend.sh
```

test:
```console
kubectl port-forward --namespace=ingress-nginx service/ingress-nginx-controller 8090:80
bash test.sh
```

FROM golang:1.20 as build

RUN mkdir -p /opt/app
COPY server.go /opt/app
COPY go.mod /opt/app
COPY go.sum /opt/app

WORKDIR "/opt/app"

RUN go build

FROM gcr.io/distroless/static-debian11

WORKDIR /
COPY --from=build /opt/app/service-b /service
EXPOSE 80
CMD ["/service", "run"]

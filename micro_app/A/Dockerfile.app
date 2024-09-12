FROM golang:1.23.1 as build

RUN mkdir -p /opt/app
COPY server.go /opt/app
COPY go.mod /opt/app
COPY go.sum /opt/app

WORKDIR "/opt/app"

RUN go build

FROM gcr.io/distroless/base-debian10

WORKDIR /
COPY --from=build /opt/app/service /service
EXPOSE 80
CMD ["/service", "run"]

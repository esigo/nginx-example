FROM golang:1.19 as build

RUN mkdir -p /opt/app
COPY server.go /opt/app

WORKDIR "/opt/app"

RUN go mod init service \
&& go mod tidy \
&& go build

FROM gcr.io/distroless/base-debian10

WORKDIR /
COPY --from=build /opt/app/service /service
EXPOSE 80
CMD ["/service", "run"]

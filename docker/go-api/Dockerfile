ARG GO_VERSION=1.17

FROM golang:${GO_VERSION} AS build

WORKDIR /app
COPY . .

RUN GOOS=linux CGO_ENABLED=0 go build ./cmd/api

FROM alpine:3.15.4

COPY --from=build /app/conf /conf
COPY --from=build /app/api /server

ENTRYPOINT ["/server", "/conf/config.prod.json"]
ARG GO_VERSION=1.17

FROM golang:${GO_VERSION} AS build

WORKDIR /app
COPY . .

RUN go mod download
RUN GOOS=linux CGO_ENABLED=0 go build ./cmd/mailer

FROM alpine:latest

COPY --from=build /app/conf /conf
COPY --from=build /app/mailer /mailer-worker

ENTRYPOINT ["/mailer-worker", "/conf/config.prod.json"]
ARG GO_VERSION=1.17

FROM golang:${GO_VERSION} AS build

WORKDIR /app
COPY . .
#RUN go mod download

RUN go build ./cmd/api
RUN go build ./cmd/mailer


FROM gcr.io/distroless/base-debian10

COPY --from=build /app/conf /conf
COPY --from=build /app/api /server
COPY --from=build /app/mailer /mailer-worker

ENTRYPOINT ["/server", "/conf/config.prod.json", "&&", "/mailer-worker", "/conf/config.prod.json"]
ARG GO_VERSION=1.16

FROM golang:${GO_VERSION} AS build

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o server \
    && go test -v \
    && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.41.1 \
    && golangci-lint run -v -c ./.golangci.yaml ./...


FROM gcr.io/distroless/base-debian10

COPY --from=build /app/conf /conf
COPY --from=build /app/server /server

ENTRYPOINT ["/server"]
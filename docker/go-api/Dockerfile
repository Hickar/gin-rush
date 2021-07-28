ARG GO_VERSION=1.16

FROM golang:${GO_VERSION} AS build

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o server && go test -v


FROM gcr.io/distroless/base-debian10

#WORKDIR /
COPY --from=build /app/conf /conf
COPY --from=build /app/server /server

ENTRYPOINT ["/server"]
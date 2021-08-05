ARG GO_VERSION=1.16

FROM golang:${GO_VERSION} AS build

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o server


FROM gcr.io/distroless/base-debian10

COPY --from=build /app/conf /conf
COPY --from=build /app/server /server

ENTRYPOINT ["/server"]
FROM golang:1.23.1 AS build

WORKDIR /app
COPY platform/api/ .
RUN go mod download && go mod tidy
RUN go build -v -o /usr/local/bin/api

FROM golang:1.23.1
COPY --from=build /usr/local/bin/api /usr/local/bin/api
ENTRYPOINT ["/usr/local/bin/api"]
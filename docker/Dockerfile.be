# Build stage
FROM golang:1.23.1 AS build
WORKDIR /app
COPY platform/api .
RUN go mod download && go mod tidy
# Change working directory to where the Go files are
WORKDIR /app/api
# Build the operator
RUN go build -v -o /usr/local/bin/api

# Final stage
FROM golang:1.23.1
COPY --from=build /usr/local/bin/api /usr/local/bin/api
ENTRYPOINT ["backend"]
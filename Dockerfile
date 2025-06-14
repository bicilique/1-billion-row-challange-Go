# syntax=docker/dockerfile:1

# Test stage
FROM builder AS tester
RUN go test ./... -v

# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY src/go.mod src/go.sum ./
RUN go mod download
COPY src/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o app main.go


# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .
COPY assets/ ./assets/
EXPOSE 8080
CMD ["./app"]

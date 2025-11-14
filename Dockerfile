# Build stage
FROM golang:1.22.11-alpine AS builder
WORKDIR /app

# Install git for dependency fetching
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /url-shortener ./cmd/main.go

# Final stage
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /url-shortener /url-shortener
EXPOSE 8080
ENTRYPOINT ["/url-shortener"]

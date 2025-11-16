# Build stage
FROM golang:1.22.11-alpine AS builder
WORKDIR /app

# Install git for dependency fetching
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /e-commerce ./cmd/main.go

# Final stage
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /e-commerce /e-commerce
EXPOSE 8080
ENTRYPOINT ["/e-commerce"]

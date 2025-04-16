# Use the official Golang image as the builder
FROM golang:1.23 AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# Final image
FROM alpine:latest

WORKDIR /app

# Install certificates for TLS and logs
RUN apk add --no-cache ca-certificates tzdata

# Copy binary and required files
COPY --from=builder /app/app .

# Default shell
ENTRYPOINT ["/bin/sh"]
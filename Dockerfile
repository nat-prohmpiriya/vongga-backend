# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Run stage
FROM alpine:latest

WORKDIR /app

# Install CA certificates
RUN apk --no-cache add ca-certificates

# Copy binary and config from builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env* ./

# Expose port
EXPOSE 4000

# Command to run
CMD ["./main"]

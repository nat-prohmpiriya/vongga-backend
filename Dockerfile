# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Run stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env .

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]

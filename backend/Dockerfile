FROM golang:1.19-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o server server.go

# Use a smaller image for the final container
FROM alpine:latest

# Install required dependencies for SQLite
RUN apk add --no-cache libc6-compat

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/server .
COPY --from=builder /app/pkg/db/migrations ./pkg/db/migrations

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./server"]

FROM golang:1.27.2-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o server server.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server ./server
COPY --from=builder /app/pkg/db/migrations ./pkg/db/migrations

# Runtime data is mounted by Docker Compose.
RUN mkdir -p /app/data /app/uploads

EXPOSE 8080
CMD ["./server"]

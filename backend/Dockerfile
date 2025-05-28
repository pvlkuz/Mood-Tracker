# Stage 1: Build the Go application and migrate CLI
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server main.go

# Stage 2: Create a runtime image
FROM alpine:3.16

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/server ./server
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

CMD ["sh", "-c", "./server"]

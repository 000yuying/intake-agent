# Dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bin/intake-agent cmd/intake-agent/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/bin/intake-agent .
COPY configs/ configs/
EXPOSE 8080
ENTRYPOINT ["./intake-agent", "--config", "configs/config.yaml"]

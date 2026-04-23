# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o aigo ./cmd/aigo

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/aigo /usr/local/bin/aigo

# Create data directory
RUN mkdir -p /root/.aigo

EXPOSE 9090 3100

ENTRYPOINT ["aigo"]
CMD ["start"]

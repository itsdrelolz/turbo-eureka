FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/worker ./cmd/workers

FROM alpine:latest

WORKDIR /root/


COPY --from=builder /bin/api .
COPY --from=builder /bin/worker .
COPY --from=builder /app/.env.example .env

EXPOSE 8080

CMD ["./api"]

# === Build Stage ===
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build the Worker binary
RUN CGO_ENABLED=0 go build -o /worker ./cmd/workers/main.go

# === Final Stage ===
FROM alpine:latest
COPY --from=builder /worker /worker

# Command to run the application
CMD ["/worker"]

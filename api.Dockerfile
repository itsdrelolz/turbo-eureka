FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the API binary.
# CGO_ENABLED=0 creates a static binary
# -o /api specifies the output file path
RUN CGO_ENABLED=0 go build -o /api ./cmd/api/main.go

FROM alpine:latest

COPY --from=builder /api /api

EXPOSE 8080

# Command to run the application
CMD ["/api"]

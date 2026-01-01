# Stage 1: Build the binary
FROM golang:1.25-alpine AS builder 

# Install SSL certificates (required for your CBA HTTPS requests)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bot-server .

# Stage 2: Final lightweight image
FROM scratch

# Copy SSL certs and the binary from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bot-server /bot-server

# App Runner usually expects traffic on a port (8080 in your main.go)
EXPOSE 8080

ENTRYPOINT ["/bot-server"]
# Build stage
FROM golang:1.21-alpine AS builder

# Set GOPROXY to use Chinese mirror
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o linux-process-exporter

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/linux-process-exporter .

# Expose metrics port
EXPOSE 9113

# Run the application
CMD ["./linux-process-exporter"]
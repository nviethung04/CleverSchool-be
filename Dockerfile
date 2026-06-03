FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git curl

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install buf
RUN curl -sSL "https://cdn-dev.xlms.vn/buf/buf-Linux-x86_64" --retry 5 --retry-delay 5 -o ./buf && \
    chmod +x ./buf

# Copy buf config files
COPY buf.yaml buf.gen.yaml ./

# Copy all source code including docs directory with embedded files
COPY . .

# Generate protobuf files using buf
RUN ./buf generate

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myapp .

# Second stage - create minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/myapp .

# Copy necessary directories for runtime
COPY --from=builder /app/database ./database
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/power_point ./power_point
COPY --from=builder /app/i18n ./i18n

# Make the binary executable
RUN chmod +x ./myapp

EXPOSE 8080

CMD ["./myapp"]

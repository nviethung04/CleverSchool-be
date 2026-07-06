FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git curl

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install buf
RUN curl -fL "https://github.com/bufbuild/buf/releases/latest/download/buf-Linux-x86_64" --retry 5 --retry-delay 5 -o /usr/local/bin/buf && \
    chmod +x /usr/local/bin/buf && \
    buf --version

# Copy buf config files
COPY buf.yaml buf.gen.yaml ./

# Copy all source code including docs directory with embedded files
COPY . .

# Generate protobuf files using buf
RUN buf generate

# Build the application and seeder CLI
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myapp . && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o seeder ./database/seeder

# Second stage - create minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /app/myapp .
COPY --from=builder /app/seeder .

# Copy necessary directories for runtime
COPY --from=builder /app/database ./database
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/power_point ./power_point
COPY --from=builder /app/i18n ./i18n
COPY docker-entrypoint.sh ./docker-entrypoint.sh

# Make the binary executable
RUN chmod +x ./myapp ./seeder ./docker-entrypoint.sh

EXPOSE 8080

CMD ["./docker-entrypoint.sh"]

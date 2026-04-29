# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first for better caching.
COPY go.mod go.sum ./
RUN go mod download

# Copy source code.
COPY . .

# Build the binary with CGO disabled (using pure-Go SQLite driver).
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# Stage 2: Runtime
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from the builder stage.
COPY --from=builder /server /app/server

# Create a directory for the SQLite database.
RUN mkdir -p /app/data

EXPOSE 8080

ENV DATABASE_PATH=/app/data/workouts.db
ENV PORT=8080

ENTRYPOINT ["/app/server"]

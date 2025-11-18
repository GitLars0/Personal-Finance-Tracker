# Stage 1: Build Go backend
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /app/backend
RUN go mod download

COPY backend/ ./ 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/Personal-Finance-Tracker

# Stage 2: Build React frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package*.json ./frontend/
WORKDIR /app/frontend
RUN npm install --legacy-peer-deps
COPY frontend/ ./ 
RUN npm run build

# Stage 3: Final runtime image
FROM alpine:latest
WORKDIR /app

# Install wget for health checks and ca-certificates for HTTPS/TLS
RUN apk --no-cache add ca-certificates wget

# Copy backend binary
COPY --from=builder /app/Personal-Finance-Tracker ./Personal-Finance-Tracker

# Copy frontend build output
COPY --from=frontend-builder /app/frontend/build ./frontend/build

# Create non-root user for security
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

EXPOSE 8080

# Add health check at Docker level (optional, already in docker-compose)
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./Personal-Finance-Tracker"]
# Stage 1: Build Frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /build
COPY frontend/package*.json ./
RUN npm ci --only=production --ignore-scripts
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.25-alpine AS backend-builder

WORKDIR /build
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o booknest-server ./cmd/server

# Stage 3: Final Production Image
FROM alpine:3.22

# Install runtime dependencies
RUN apk add --no-cache ca-certificates su-exec tzdata

# Create app user and group
RUN addgroup -g 1000 appgroup && \
    adduser -D -H -u 1000 -G appgroup appuser

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /build/booknest-server /app/booknest-server

# Copy frontend build artifacts
COPY --from=frontend-builder /build/dist /app/frontend/dist

# Copy docker entrypoint
COPY backend/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# Set environment variables
ENV PUID=1000 \
    PGID=1000 \
    DATA_DIR=/data \
    BOOKS_DIR=/data/books \
    COVERS_DIR=/data/covers \
    ASSETS_DIR=/data/assets \
    TEMP_DIR=/data/temp \
    FRONTEND_DIST_DIR=/app/frontend/dist \
    HTTP_ADDR=:8080 \
    APP_ENV=production

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/booknest-server"]

FROM node:22-alpine AS frontend-build

WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.25-alpine AS backend-build

WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/booknest-server ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache su-exec \
	&& mkdir -p /app/frontend/dist /data/books /data/covers /data/assets /data/temp

WORKDIR /app
COPY --from=backend-build /out/booknest-server /app/booknest-server
COPY --from=frontend-build /src/frontend/dist /app/frontend/dist
COPY backend/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

ENV APP_ENV=production \
	HTTP_ADDR=:8080 \
	DATA_DIR=/data \
	BOOKS_DIR=/data/books \
	COVERS_DIR=/data/covers \
	ASSETS_DIR=/data/assets \
	TEMP_DIR=/data/temp \
	FRONTEND_DIST_DIR=/app/frontend/dist \
	PUID=1000 \
	PGID=1000

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/booknest-server"]

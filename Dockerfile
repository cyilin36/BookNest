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

RUN addgroup -S app && adduser -S -D -H -u 10001 -G app appuser \
	&& mkdir -p /app/frontend/dist /data/books /data/covers /data/temp \
	&& chown -R appuser:app /app /data

WORKDIR /app
COPY --from=backend-build /out/booknest-server /app/booknest-server
COPY --from=frontend-build /src/frontend/dist /app/frontend/dist

ENV APP_ENV=production \
	HTTP_ADDR=:8080 \
	DATA_DIR=/data \
	FRONTEND_DIST_DIR=/app/frontend/dist

USER appuser
EXPOSE 8080

ENTRYPOINT ["/app/booknest-server"]

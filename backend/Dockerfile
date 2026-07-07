FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/booknest-server ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache su-exec
RUN addgroup -g 1000 appgroup && adduser -D -H -u 1000 -G appgroup appuser
WORKDIR /app
COPY --from=build /out/booknest-server /app/booknest-server
COPY --from=build /src/assets /app/assets
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

ENV PUID=1000
ENV PGID=1000
ENV DATA_DIR=/data
ENV BOOKS_DIR=/data/books
ENV COVERS_DIR=/data/covers
ENV ASSETS_DIR=/data/assets
ENV TEMP_DIR=/data/temp

EXPOSE 8080
ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/booknest-server"]

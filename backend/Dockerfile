FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/book-reader-server ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
WORKDIR /app
COPY --from=build /out/book-reader-server /app/book-reader-server
USER appuser

EXPOSE 8080
ENTRYPOINT ["/app/book-reader-server"]

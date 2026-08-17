FROM golang:1.23-alpine AS build

WORKDIR /src
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/blog-api ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=build /out/blog-api /app/blog-api
COPY migrations /app/migrations
COPY configs /app/configs

ENV APP_PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=5s --timeout=3s --start-period=10s --retries=24 \
  CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1

CMD ["./blog-api"]

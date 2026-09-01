# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/n07 ./cmd/nexus
RUN GOBIN=/out go install github.com/storacha/guppy@v0.7.0

FROM alpine:3.22
RUN addgroup -S n07 && adduser -S -G n07 n07 \
    && apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/n07 /app/n07
COPY --from=builder /out/guppy /usr/local/bin/guppy
RUN mkdir -p /var/lib/n07/storacha && chown -R n07:n07 /var/lib/n07
USER n07
ENV N07_HTTP_ADDR=:8080 \
    STORACHA_GUPPY_BIN=/usr/local/bin/guppy \
    STORACHA_DATA_DIR=/var/lib/n07/storacha
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/health >/dev/null || exit 1
ENTRYPOINT ["/app/n07"]

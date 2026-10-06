# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN set -eu; \
    count=0; first=1; \
    { \
      echo '{'; \
      echo '  "schema_version": "soul.external.runtime-attestation.v1",'; \
      echo '  "generated_from_materialized_git": true,'; \
      echo '  "repositories": {'; \
      for d in integrations/external/*; do \
        test -d "$d"; \
        id="$(basename "$d")"; \
        sha="$(git -C "$d" rev-parse HEAD)"; \
        gitlink="$(git ls-tree HEAD -- "$d" | awk '{print $3}')"; \
        test -n "$sha" && test "$gitlink" = "$sha"; \
        if [ "$first" -eq 0 ]; then printf ',\n'; fi; \
        printf '    "%s": "%s"' "$id" "$sha"; \
        first=0; count=$((count+1)); \
      done; \
      echo; \
      echo '  }'; \
      echo '}'; \
      test "$count" -eq 22; \
    } > config/soul-external-runtime-attestation.json
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/n07 ./cmd/nexus
RUN set -eux; \
    for attempt in 1 2 3 4 5; do \
      if GOBIN=/out go install github.com/storacha/guppy@v0.7.0; then \
        break; \
      fi; \
      if [ "$attempt" -eq 5 ]; then \
        echo "guppy install failed after 5 attempts" >&2; \
        exit 1; \
      fi; \
      sleep $((attempt * 3)); \
    done

FROM alpine:3.22
RUN addgroup -S n07 && adduser -S -G n07 n07 \
    && apk add --no-cache ca-certificates tzdata python3
WORKDIR /app
COPY --from=builder /out/n07 /app/n07
COPY --from=builder /out/guppy /usr/local/bin/guppy
COPY --from=builder /src/integrations/external /app/integrations/external
COPY --from=builder /src/integrations/external-adapters.json /app/integrations/external-adapters.json
COPY --from=builder /src/integrations/external-capabilities.json /app/integrations/external-capabilities.json
COPY --from=builder /src/scripts/external-capability-adapter.py /app/scripts/external-capability-adapter.py
COPY --from=builder /src/scripts/crewai_runner.py /app/scripts/crewai_runner.py
COPY --from=builder /src/scripts/metagpt_runner.py /app/scripts/metagpt_runner.py
COPY --from=builder /src/config/soul-external-runtime-attestation.json /app/config/soul-external-runtime-attestation.json
RUN mkdir -p /var/lib/n07/storacha /var/lib/n07/external-io && chown -R n07:n07 /app /var/lib/n07
USER n07
ENV N07_HTTP_ADDR=:8080 \
    SOUL_EXTERNAL_PYTHON=/usr/bin/python3 \
    SOUL_EXTERNAL_IO_ROOT=/var/lib/n07/external-io \
    STORACHA_GUPPY_BIN=/usr/local/bin/guppy \
    STORACHA_DATA_DIR=/var/lib/n07/storacha
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/health >/dev/null || exit 1
ENTRYPOINT ["/app/n07"]

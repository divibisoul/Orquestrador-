FROM node:20-alpine
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY package.json gateway.mjs ./
RUN addgroup -S soul && adduser -S -G soul soul \
    && mkdir -p /var/lib/soul-agent-arsenal \
    && chown -R soul:soul /app /var/lib/soul-agent-arsenal
USER soul
ENV AGENT_ARSENAL_ROOT=/var/lib/soul-agent-arsenal \
    AGENT_ARSENAL_HOST=0.0.0.0 \
    AGENT_ARSENAL_PORT=8090 \
    AGENT_ARSENAL_EXECUTE=false \
    AGENT_ARSENAL_COMMAND_TIMEOUT_MS=20000 \
    AGENT_ARSENAL_MAX_RESOLVE_BYTES=1048576
EXPOSE 8090
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
  CMD node gateway.mjs --healthcheck
ENTRYPOINT ["node","gateway.mjs"]

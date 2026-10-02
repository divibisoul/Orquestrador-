FROM node:20-bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY package.json gateway.mjs ./
RUN node --check gateway.mjs

ENV AGENT_ARSENAL_ROOT=/var/lib/soul-agent-arsenal
ENV AGENT_ARSENAL_PORT=8090
ENV AGENT_ARSENAL_EXECUTE=false

VOLUME ["/var/lib/soul-agent-arsenal"]
EXPOSE 8090

CMD ["node", "gateway.mjs"]

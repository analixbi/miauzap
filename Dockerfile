FROM golang:1.25-bookworm AS builder
ENV GOTOOLCHAIN=auto

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    g++ \
    pkg-config \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=1
RUN go build -o miauzap

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    netcat-openbsd \
    postgresql-client \
    openssl \
    curl \
    ffmpeg \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

ENV TZ="America/Sao_Paulo"
# Create a non-privileged user
RUN groupadd -r miauzap && useradd -r -g miauzap miauzap

WORKDIR /app

COPY --from=builder /app/miauzap         /app/
COPY --from=builder /app/static         /app/static/

RUN chmod +x /app/miauzap && \
    chmod -R 755 /app && \
    chown -R miauzap:miauzap /app

USER miauzap

ENTRYPOINT ["/app/miauzap", "--logtype=console", "--color=true"]

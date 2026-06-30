# ---------- Stage 1: build ----------
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Dependências primeiro — aproveita o cache de camadas do Docker.
COPY go.mod go.sum* ./
RUN go mod download

# Código-fonte.
COPY . .

# Binário estático (sem CGO) para rodar em imagem mínima.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /build/barbercentral-backend \
    ./cmd/api

# ---------- Stage 2: runtime ----------
FROM alpine:3.20

# curl para o HEALTHCHECK; ca-certificates e tzdata para TLS e fuso.
RUN apk add --no-cache ca-certificates tzdata curl

# Usuário não-root.
RUN addgroup -g 1000 app && adduser -D -u 1000 -G app app

WORKDIR /app

# Binário e migrations.
COPY --from=builder /build/barbercentral-backend .
COPY --from=builder /build/migrations ./migrations

RUN chown -R app:app /app
USER app

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

CMD ["./barbercentral-backend"]

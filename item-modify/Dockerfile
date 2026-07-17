# ── Stage 1: build ──────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Cache dependencies first
COPY api/go.mod api/go.sum ./
RUN go mod download

# Copy source
COPY api/ ./

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /app/server .

# ── Stage 2: runtime ────────────────────────────────────────────
FROM scratch

# Copy CA certificates so HTTPS calls to IBM Cloud work
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /app/server /server

# Copy the UI directory (served as static files at /)
COPY ui/ /ui/

# Code Engine injects PORT; default to 8080
ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/server"]

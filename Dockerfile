# ═══ Stage 1: Build ═══
FROM golang:1.24-alpine AS builder
ENV GOTOOLCHAIN=auto
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /seed ./cmd/seed

# ═══ Stage 2: Runtime ═══
FROM golang:1.24-alpine
RUN apk --no-cache add ca-certificates tzdata openssh-client && \
    adduser -D -g '' appuser && mkdir -p /app/data && chown appuser /app/data
WORKDIR /app
COPY --from=builder /server /app/server
COPY --from=builder /seed /app/seed
COPY internal/templates /app/internal/templates
COPY internal/static /app/internal/static
COPY migrations /app/migrations
USER appuser
EXPOSE 8080
ENV TEMPLATE_DIR=/templates
ENV STATIC_DIR=/static
CMD ["/app/server"]
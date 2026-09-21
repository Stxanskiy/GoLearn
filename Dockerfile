# ═══ Stage 1: build ═══
FROM golang:1.24-alpine AS builder
ENV GOTOOLCHAIN=auto
WORKDIR /app
# Dependencies first: a source-only change then reuses this layer.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /seed ./cmd/seed

# ═══ Stage 2: runtime ═══
# alpine rather than the Go image: the toolchain is a build-time need, and
# carrying it into the runtime was most of the old image's size.
FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata openssh-client && \
    adduser -D -g '' appuser && mkdir -p /app/data && chown appuser /app/data
WORKDIR /app
COPY --from=builder /server /app/server
COPY --from=builder /seed /app/seed
# The service renders no pages, so no templates and no static assets. What it
# still needs at runtime: the migrations it applies on boot, and the OpenAPI
# contract it serves at /docs — the only description of this API now.
COPY migrations /app/migrations
COPY api/openapi.yaml /app/api/openapi.yaml
USER appuser
EXPOSE 8080
CMD ["/app/server"]

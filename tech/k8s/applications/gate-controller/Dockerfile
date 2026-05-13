FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o gate-controller \
    ./tech/k8s/applications/gate-controller


FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata wget && \
    adduser -S appuser

WORKDIR /app

COPY --from=builder /build/gate-controller .

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health-check || exit 1

USER appuser

ENTRYPOINT ["/app/gate-controller"]

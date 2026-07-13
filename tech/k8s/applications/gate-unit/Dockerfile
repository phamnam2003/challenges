FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o gate-unit \
    ./gate-unit


FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -S appuser

WORKDIR /app

COPY --from=builder /build/gate-unit .

EXPOSE 8081

USER appuser

ENTRYPOINT ["/app/gate-unit"]

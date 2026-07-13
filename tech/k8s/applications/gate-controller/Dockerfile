FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o gate-controller \
    ./gate-controller


FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -S appuser

WORKDIR /app

COPY --from=builder /build/gate-controller .

EXPOSE 8080

USER appuser

ENTRYPOINT ["/app/gate-controller"]

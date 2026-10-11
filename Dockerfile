# ── Stage 1: Build ──
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /api ./cmd/api

# ── Stage 2: Runtime ──
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Jakarta

WORKDIR /app

COPY --from=builder /api .
COPY migrations/ ./migrations/

RUN mkdir -p uploads/attendance

EXPOSE 8090

CMD ["./api"]

FROM golang:1.21-alpine AS builder

ENV GO111MODULE=on

WORKDIR /app

COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o star-mining-server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/star-mining-server .

ENV PORT=8080
ENV REDIS_ADDR=redis:6379
ENV GIN_MODE=release
ENV STATIC_DIR=

EXPOSE 8080

CMD ["./star-mining-server"]

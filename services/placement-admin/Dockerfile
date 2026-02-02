FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o service ./cmd/service

FROM alpine:3.20
WORKDIR /app
ENV PORT=8080
COPY --from=builder /app/service ./service
CMD ["./service"]

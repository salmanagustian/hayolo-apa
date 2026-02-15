FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o app ./cmd

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/app .

ENV PORT=8080
CMD ["./app"]

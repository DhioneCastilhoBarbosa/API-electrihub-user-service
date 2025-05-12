# Etapa de build
FROM golang:1.23.5 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/server

# Etapa final
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8087

CMD ["./main"]

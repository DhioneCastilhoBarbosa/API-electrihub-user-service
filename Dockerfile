# Etapa de build
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o app ./cmd/server

# Etapa de produção
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/app .

EXPOSE 8087
CMD ["./app"]

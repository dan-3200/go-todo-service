# --- Etapa de Construção
FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd

# --- Etapa de Execução
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8200

CMD ["./server"]

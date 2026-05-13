FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/log-parser ./cmd/app

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /bin/log-parser /app/log-parser
COPY internal/db/migrations.sql /app/internal/db/migrations.sql

EXPOSE 8080

CMD ["/app/log-parser"]

FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY services/chat/go.mod /src/services/chat/go.mod
COPY services/pkg/go.mod /src/services/pkg/go.mod

WORKDIR /src/services/chat
RUN go mod download

COPY services/chat /src/services/chat
COPY services/pkg /src/services/pkg
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/service ./cmd/main.go

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /out/service /app/service
RUN touch /app/.env

EXPOSE 8089
ENTRYPOINT ["/app/service"]

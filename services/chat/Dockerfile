FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY chat/go.mod /src/chat/go.mod
COPY pkg/go.mod /src/pkg/go.mod

WORKDIR /src/chat
RUN go mod download

COPY chat /src/chat
COPY pkg /src/pkg
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/service ./cmd/main.go

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /out/service /app/service
RUN touch /app/.env

EXPOSE 8089
ENTRYPOINT ["/app/service"]

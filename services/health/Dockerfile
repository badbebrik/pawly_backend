FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY health/go.mod health/go.sum /src/health/
COPY pkg/go.mod pkg/go.sum /src/pkg/

WORKDIR /src/health
RUN go mod download

COPY health /src/health
COPY pkg /src/pkg
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/service ./cmd/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /out/service /app/service
RUN touch /app/.env

EXPOSE 8088
ENTRYPOINT ["/app/service"]

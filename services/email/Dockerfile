FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY email/go.mod email/go.sum /src/email/
COPY pkg/go.mod pkg/go.sum /src/pkg/

WORKDIR /src/email
RUN go mod download

COPY email /src/email
COPY pkg /src/pkg
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/service ./cmd/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /out/service /app/service
COPY --from=builder /src/email/templates /app/templates
RUN touch /app/.env

EXPOSE 8084
ENTRYPOINT ["/app/service"]

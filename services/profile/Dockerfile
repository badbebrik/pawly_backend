FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY profile/go.mod profile/go.sum /src/profile/
COPY pkg/go.mod pkg/go.sum /src/pkg/

WORKDIR /src/profile
RUN go mod download

COPY profile /src/profile
COPY pkg /src/pkg
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/service ./cmd/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /out/service /app/service
RUN touch /app/.env

EXPOSE 8086 50058
ENTRYPOINT ["/app/service"]

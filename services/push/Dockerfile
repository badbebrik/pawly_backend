FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY push/go.mod push/go.sum /src/push/
COPY pkg/go.mod pkg/go.sum /src/pkg/

WORKDIR /src/push
RUN go mod download

COPY push /src/push
COPY pkg /src/pkg
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/service ./cmd/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /out/service /app/service

EXPOSE 8090
ENTRYPOINT ["/app/service"]

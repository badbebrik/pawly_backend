FROM golang:1.24-alpine AS builder
WORKDIR /src

# profile depends on local module "file" via replace directive
COPY profile/go.mod profile/go.sum /src/profile/
COPY file/go.mod file/go.sum /src/file/

WORKDIR /src/profile
RUN go mod download

COPY profile /src/profile
COPY file /src/file
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/service ./cmd/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /out/service /app/service
RUN touch /app/.env

EXPOSE 8086
ENTRYPOINT ["/app/service"]

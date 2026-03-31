FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY services/chat/go.mod /src/services/chat/go.mod
COPY services/acl/go.mod /src/services/acl/go.mod
COPY services/profile/go.mod /src/services/profile/go.mod
COPY services/pet/go.mod /src/services/pet/go.mod
COPY services/pkg/go.mod /src/services/pkg/go.mod

WORKDIR /src/services/chat
RUN go mod download

COPY services/chat /src/services/chat
COPY services/acl /src/services/acl
COPY services/profile /src/services/profile
COPY services/pet /src/services/pet
COPY services/pkg /src/services/pkg
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/service ./cmd/main.go

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /out/service /app/service
RUN touch /app/.env

EXPOSE 8089
ENTRYPOINT ["/app/service"]

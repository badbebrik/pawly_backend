FROM golang:1.24-alpine

RUN apk add --no-cache bash ca-certificates git && \
    go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

WORKDIR /work

COPY deploy/migrate/run-migrations.sh /usr/local/bin/run-migrations
RUN chmod +x /usr/local/bin/run-migrations

ENTRYPOINT ["run-migrations"]

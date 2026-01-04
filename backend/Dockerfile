FROM golang:1.25.4-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git binutils postgresql-client
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -trimpath \
    -o server \
    ./cmd/server/main.go ./cmd/server/wire_gen.go

RUN strip server

# FROM scratch AS runtime # TODO: Put back in production
FROM debian:bookworm-slim AS runtime

WORKDIR /app

COPY --from=builder /build/server /app/server

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080

ENTRYPOINT ["/app/server"]

FROM builder AS migration

WORKDIR /app

RUN apk add --no-cache postgresql-client ca-certificates && \
    cp /go/bin/migrate /usr/local/bin/migrate && \
    cp -r /build/migrations /app/migrations && \
    cp /build/scripts/seed_reelstrips/*.csv /app/

RUN go install github.com/google/wire/cmd/wire@latest && \
    cd /build/scripts/seed_assets && wire && go build -o /app/seed_assets . && \
    cd /build/scripts/seed_reelstrips && wire && go build -o /app/seed .

ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["/bin/sh"]

FROM golang:1.22 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY /cmd ./cmd
COPY /internal ./internal
COPY /pkg ./pkg

RUN go build -v -o ./bin/server ./cmd/server
RUN go build -v -o ./bin/client ./cmd/client

FROM alpine AS server
WORKDIR /app
COPY --from=builder /build/bin/server ./
RUN apk add libc6-compat
CMD ["/app/server"]

FROM alpine AS client
WORKDIR /app
COPY --from=builder /build/bin/client ./
RUN apk add libc6-compat
CMD ["/app/client"]

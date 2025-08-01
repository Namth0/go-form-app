FROM golang:1.23.4-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata
RUN adduser -D -s /bin/sh appuser

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go test ./...

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main main.go

FROM alpine:3.19

RUN apk add --no-cache \
    python3 \
    py3-pip \
    bash \
    zsh \
    wget \
    ca-certificates \
    tzdata \
    && ln -sf python3 /usr/bin/python

RUN adduser -D -s /bin/sh appuser

COPY --from=builder /app/main /main
COPY --from=builder /app/cmd/server/http/web /cmd/server/http/web
COPY --from=builder /app/internal/scripts /internal/scripts

EXPOSE 8001

USER appuser

CMD ["/main"]

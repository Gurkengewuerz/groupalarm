FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY src/go.mod src/go.sum ./
RUN go mod download && go mod verify
COPY src/ .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /app/app ./...

FROM alpine:3.19

RUN apk add --no-cache ca-certificates jq tzdata
WORKDIR /app
COPY --from=builder /app/app .
COPY docker/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]

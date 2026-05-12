FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o go-whatsapp-mcp .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates && \
    addgroup -S mcp && \
    adduser -S mcp -G mcp

WORKDIR /app
COPY --from=builder /app/go-whatsapp-mcp .

RUN mkdir -p /data && chown mcp:mcp /data /app/go-whatsapp-mcp
VOLUME /data

ENV WHATSAPP_DB_PATH=/data/whatsapp.db
USER mcp

ENTRYPOINT ["./go-whatsapp-mcp"]

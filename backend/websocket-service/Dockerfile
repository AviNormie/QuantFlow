FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY shared/go.mod shared/go.sum ./shared/
COPY websocket-service/go.mod websocket-service/go.sum ./websocket-service/

WORKDIR /app/websocket-service
RUN go mod download

COPY shared/ /app/shared/
COPY websocket-service/ .

RUN CGO_ENABLED=0 go build -o /server ./cmd

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /server .

ENV PORT=8083
EXPOSE 8083

CMD ["./server"]

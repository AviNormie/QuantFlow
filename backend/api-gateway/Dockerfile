FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY shared/go.mod shared/go.sum ./shared/
COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/

WORKDIR /app/api-gateway
RUN go mod download

COPY shared/ /app/shared/
COPY api-gateway/ .

RUN CGO_ENABLED=0 go build -o /server ./cmd

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /server .

ENV PORT=8080
EXPOSE 8080

CMD ["./server"]

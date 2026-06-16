FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY shared/go.mod shared/go.sum ./shared/
COPY auth-service/go.mod auth-service/go.sum ./auth-service/

WORKDIR /app/auth-service
RUN go mod download

COPY shared/ /app/shared/
COPY auth-service/ .

RUN CGO_ENABLED=0 go build -o /server ./cmd

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /server .

ENV PORT=8084
EXPOSE 8084

CMD ["./server"]

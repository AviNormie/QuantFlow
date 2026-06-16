FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY shared/go.mod shared/go.sum ./shared/
COPY market-service/go.mod market-service/go.sum ./market-service/

WORKDIR /app/market-service
RUN go mod download

COPY shared/ /app/shared/
COPY market-service/ .

RUN CGO_ENABLED=0 go build -o /server ./cmd

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /server .

ENV PORT=8082
EXPOSE 8082

CMD ["./server"]

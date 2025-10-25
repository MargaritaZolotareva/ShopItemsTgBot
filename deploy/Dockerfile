FROM golang:1.22-alpine as builder

RUN apk add --no-cache git

WORKDIR /app
COPY .. .

RUN go mod tidy && go build -o ShopItemsTgBot cmd/server/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /build
COPY --from=builder /app/ShopItemsTgBot .
#COPY --from=builder /app/.env .


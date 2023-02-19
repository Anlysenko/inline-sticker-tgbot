# builder
FROM golang:1.20-alpine as builder

WORKDIR /

RUN apk --update upgrade && \
   apk add git && \
   rm -rf /var/cache/apk/*

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app .

FROM alpine:latest

WORKDIR /
COPY --from=builder /app .

CMD ["./app"]
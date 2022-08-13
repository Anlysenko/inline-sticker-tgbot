# builder
FROM golang:1.19-alpine as builder

WORKDIR /

RUN apk --update upgrade && \
   apk add git && \
   rm -rf /var/cache/apk/*

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app .

FROM alpine:latest 

WORKDIR /
COPY --from=builder /app .
COPY --from=builder /db.sql .

CMD ["./app"]
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY . .

RUN apk add --no-cache git

RUN go build -o main main.go
RUN apk add curl
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM alpine:3.14

WORKDIR /app

COPY --from=builder /app/main . 
COPY --from=builder /go/bin/goose ./goose
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY db/migration ./migration
RUN chmod +x /app/goose /app/start.sh /app/wait-for.sh

EXPOSE 8080

CMD ["/app/main"]
ENTRYPOINT [ "/app/start.sh", "/app/main"]
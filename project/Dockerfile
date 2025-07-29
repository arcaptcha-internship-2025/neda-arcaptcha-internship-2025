FROM golang:1.23.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o apartment

FROM alpine:latest


RUN apk add --no-cache curl


COPY --from=builder /app/apartment /app/apartment

EXPOSE 8080

CMD ["/app/apartment", "start"]
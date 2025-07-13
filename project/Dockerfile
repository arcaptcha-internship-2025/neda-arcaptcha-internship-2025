FROM golang:1.23.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o appartement

FROM scratch
COPY --from=builder /app/appartement /app/appartement

EXPOSE 8080

CMD [ "/app/apartmant start" ]
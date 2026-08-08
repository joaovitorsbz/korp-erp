FROM golang:1-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/http-server-projeto-korp .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/http-server-projeto-korp /app/http-server-projeto-korp

EXPOSE 8080
ENTRYPOINT ["/app/http-server-projeto-korp"]

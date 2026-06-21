FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api

FROM alpine:3.22

RUN adduser -D -H appuser
USER appuser

COPY --from=builder /bin/api /bin/api

EXPOSE 3000

ENTRYPOINT ["/bin/api"]

FROM m.daocloud.io/docker.io/library/golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o lite-collector .

FROM m.daocloud.io/docker.io/library/alpine:3.20

WORKDIR /app

RUN apk add --no-cache tzdata ca-certificates

COPY --from=builder /app/lite-collector .

EXPOSE 8080

CMD ["./lite-collector"]

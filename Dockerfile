FROM m.daocloud.io/docker.io/library/golang:1.26-alpine AS builder

WORKDIR /app

# Install CA certs and timezone data into the builder so we can copy them
RUN apk add --no-cache ca-certificates tzdata

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary — no libc dependency, safe to run in scratch
RUN CGO_ENABLED=0 GOOS=linux go build -o lite-collector .

# Final stage: scratch has no OS at all, just our binary + certs
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/lite-collector /app/lite-collector

EXPOSE 8080

CMD ["/app/lite-collector"]

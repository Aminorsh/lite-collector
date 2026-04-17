FROM m.daocloud.io/docker.io/library/golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO disabled keeps the binary statically linked; the runtime libraries we
# pull in below are only needed for headless chromium, not for the Go app.
RUN CGO_ENABLED=0 GOOS=linux go build -o lite-collector .

# Final stage: alpine + chromium for PDF export.
# Switched from scratch because the /jobs/:id/pdf endpoint renders reports
# via headless chromium (see services/pdf_service.go).
FROM m.daocloud.io/docker.io/library/alpine:3.20

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    chromium \
    font-noto-cjk \
    ttf-freefont

COPY --from=builder /app/lite-collector /app/lite-collector

EXPOSE 8080

CMD ["/app/lite-collector"]

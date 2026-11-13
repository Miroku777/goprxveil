FROM golang:1.27-bookworm AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -tags timetzdata -o /app/bin/app ./cmd/app

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    chromium \
    xvfb \
    xauth \
    nodejs \
    npm \
    fonts-liberation \
    fonts-unifont \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/bin/app /app/app
COPY --from=builder /app/templates /app/templates

# ENV PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
# RUN npm install -g playwright@1.64.1

# COPY ./playwright-driver /root/.cache/ms-playwright-go

COPY ./playwright-driver /root/.cache/ms-playwright-go

ENV HOME=/root
EXPOSE 8888

CMD ["sh", "-c", "xvfb-run -a -s '-screen 0 1920x1080x24' /app/app"]
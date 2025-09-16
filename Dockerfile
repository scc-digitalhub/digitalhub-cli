# ---------- build stage ----------
FROM golang:1.23.4-alpine AS builder

RUN apk add --no-cache git
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary for maximum Linux compatibility
ENV CGO_ENABLED=0 GOOS=linux

ARG MAIN_PKG=.
RUN mkdir -p /out && \
    go build -trimpath -ldflags="-s -w" -o /out/dhcli ${MAIN_PKG}

# ---------- runtime stage ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/dhcli /usr/local/bin/dhcli
RUN chmod +x /usr/local/bin/dhcli

RUN addgroup --gid 65500 dhcli
RUN adduser --uid 65500 -G dhcli -S dhcli
USER 65500

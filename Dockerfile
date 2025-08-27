# ---------- build stage ----------
FROM golang:1.23.4-alpine AS builder

# alcune dipendenze di mod richiedono git
RUN apk add --no-cache git

WORKDIR /src

# cache dipendenze
COPY go.mod go.sum ./
RUN go mod download

# sorgenti
COPY . .

# binario statico per linux (massima compatibilità)
ENV CGO_ENABLED=0 GOOS=linux

# path del package main (di default la root del modulo)
ARG MAIN_PKG=.

# compila SOLO il package main indicato
RUN mkdir -p /out && \
    go build -trimpath -ldflags="-s -w" -o /out/dhcli ${MAIN_PKG}

# ---------- runtime stage ----------
FROM alpine:3.20

# /bin/sh serve all'initContainer; CA per HTTPS
RUN apk add --no-cache ca-certificates

# metti dhcli nel PATH
COPY --from=builder /out/dhcli /usr/local/bin/dhcli
RUN chmod +x /usr/local/bin/dhcli

# niente ENTRYPOINT: l'initContainer userà /bin/sh -c "command -v dhcli"

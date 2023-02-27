FROM golang:1.20.1-alpine3.15 as builder
RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/v3.15/main/ gcc=~10.3 pkgconfig=~1.7 musl-dev=~1.2 libgit2-dev=~1.3 binutils-gold=~2.37
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY src/ src/
RUN CGO_ENABLED=1 go build -o azcagit src/main.go

FROM alpine:3.15
LABEL org.opencontainers.image.source="https://github.com/XenitAB/azcagit"
# hadolint ignore=DL3017,DL3018
RUN apk upgrade --no-cache && \
    apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/v3.15/main/ ca-certificates tini=~0.19 libgit2=~1.3

COPY --from=builder /workspace/azcagit /usr/local/bin/
WORKDIR /workspace
ENTRYPOINT [ "/sbin/tini", "--", "azcagit" ]
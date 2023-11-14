# BUILDER
FROM golang:1.21.4-bullseye as builder
WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY Makefile Makefile
COPY src/ src/

RUN make build

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-static /tini
RUN chmod +x /tini

# hadolint ignore=DL3008
RUN apt-get update && \
    apt-get -y --no-install-recommends install ca-certificates && \
    update-ca-certificates

# RUNTIME
FROM gcr.io/distroless/static-debian11:nonroot
LABEL org.opencontainers.image.source="https://github.com/XenitAB/azcagit"

WORKDIR /
COPY --from=builder /workspace/bin/azcagit /azcagit
COPY --from=builder /tini /tini
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

ENTRYPOINT [ "/tini", "--", "/azcagit"]
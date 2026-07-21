FROM --platform=$BUILDPLATFORM snowdreamtech/golang:1.26.5-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

# Set the workdir
WORKDIR /root/go/src/github.com/snowdreamtech/unistack

ENV GO111MODULE=auto

COPY . ./

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags "-s -w" -v -o /root/go/bin/unistack

# Base Image (Alpine Variant)
# Extends snowdreamtech/alpine with additional tooling and configuration.

FROM ghcr.io/snowdreamtech/alpine:3.24.1

# Build-time Configuration
ARG BUILDTIME \
    VERSION \
    REVISION \
    KEEPALIVE=0 \
    CAP_NET_BIND_SERVICE=0 \
    LANG=C.UTF-8 \
    UMASK=022 \
    DEBUG=false \
    PASSWORDLESS_SUDO=false \
    PGID=0 \
    PUID=0 \
    USER=root \
    WORKDIR=/root

# Standard OCI Metadata (https://github.com/opencontainers/image-spec)
LABEL org.opencontainers.image.authors="Snowdream Tech" \
    org.opencontainers.image.title="UniStack Image Based On Alpine" \
    org.opencontainers.image.description="Docker Images for UniStack on Alpine. (i386, amd64, arm32v6, arm32v7, arm64, ppc64le, riscv64, s390x)" \
    org.opencontainers.image.documentation="https://hub.docker.com/r/snowdreamtech/unistack" \
    org.opencontainers.image.base.name="snowdreamtech/unistack:alpine" \
    org.opencontainers.image.licenses="MIT" \
    org.opencontainers.image.source="https://github.com/snowdreamtech/unistack" \
    org.opencontainers.image.vendor="Snowdream Tech" \
    org.opencontainers.image.created="${BUILDTIME}" \
    org.opencontainers.image.revision="${REVISION}" \
    org.opencontainers.image.url="https://github.com/snowdreamtech/unistack"

# x-release-please-start-version
LABEL org.opencontainers.image.version="${VERSION:-0.0.3}"
# x-release-please-end

# Runtime Environment Configuration
ENV KEEPALIVE=${KEEPALIVE} \
    CAP_NET_BIND_SERVICE=${CAP_NET_BIND_SERVICE} \
    LANG=${LANG} \
    UMASK=${UMASK} \
    DEBUG=${DEBUG} \
    PASSWORDLESS_SUDO=${PASSWORDLESS_SUDO} \
    PGID=${PGID} \
    PUID=${PUID} \
    USER=${USER} \
    WORKDIR=${WORKDIR}

COPY --from=builder /root/go/bin/unistack /usr/local/bin/

RUN chmod +x /usr/local/bin/unistack

# Container Orchestration Files
COPY docker/alpine/entrypoint.d /usr/local/bin/entrypoint.d

RUN chmod +x /usr/local/bin/entrypoint.d/*

# Standard Healthcheck (Verification of core system responsiveness)
HEALTHCHECK NONE

# Process Execution
# hadolint ignore=DL3002
USER root
WORKDIR /root
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["/bin/sh"]

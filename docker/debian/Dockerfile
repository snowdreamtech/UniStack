FROM --platform=$BUILDPLATFORM snowdreamtech/golang:1.24.2-debian AS builder

ARG TARGETOS
ARG TARGETARCH

# Set the workdir
WORKDIR /root/go/src/github.com/snowdreamtech/unigo

ENV GO111MODULE=auto

COPY . ./

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags "-s -w" -v -o /root/go/bin/unigo

# Base Image (Debian Variant)
# Extends snowdreamtech/debian with additional tooling and configuration.

FROM ghcr.io/snowdreamtech/debian:13.5.0

# Build-time Configuration
ARG BUILDTIME \
    VERSION \
    REVISION \
    DEBIAN_FRONTEND=noninteractive \
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
    org.opencontainers.image.title="UniGo Image Based On Debian" \
    org.opencontainers.image.description="Docker Images for UniGo on Debian. (i386, amd64, arm32v5, arm32v7, arm64, mips64le, ppc64le, s390x)" \
    org.opencontainers.image.documentation="https://hub.docker.com/r/snowdreamtech/unigo" \
    org.opencontainers.image.base.name="snowdreamtech/unigo:debian" \
    org.opencontainers.image.licenses="MIT" \
    org.opencontainers.image.source="https://github.com/snowdreamtech/unigo" \
    org.opencontainers.image.vendor="Snowdream Tech" \
    org.opencontainers.image.version="${VERSION:-13.5.0}" \
    org.opencontainers.image.created="${BUILDTIME}" \
    org.opencontainers.image.revision="${REVISION}" \
    org.opencontainers.image.url="https://github.com/snowdreamtech/unigo"

# Runtime Environment Configuration
ENV DEBIAN_FRONTEND=${DEBIAN_FRONTEND} \
    KEEPALIVE=${KEEPALIVE} \
    CAP_NET_BIND_SERVICE=${CAP_NET_BIND_SERVICE} \
    LANG=${LANG} \
    UMASK=${UMASK} \
    DEBUG=${DEBUG} \
    PASSWORDLESS_SUDO=${PASSWORDLESS_SUDO} \
    PGID=${PGID} \
    PUID=${PUID} \
    USER=${USER} \
    WORKDIR=${WORKDIR}

COPY --from=builder /root/go/bin/unigo /usr/local/bin/

RUN chmod +x /usr/local/bin/unigo

# Container Orchestration Files
COPY entrypoint.d /usr/local/bin/entrypoint.d

RUN chmod +x /usr/local/bin/entrypoint.d/*

# Standard Healthcheck (Verification of core system responsiveness)
HEALTHCHECK NONE

# Process Execution
USER root
WORKDIR /root
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["/bin/bash"]

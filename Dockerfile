# syntax = docker/dockerfile:1.2

FROM golang:1.19.1-bullseye AS build

WORKDIR /workspace

COPY . ./

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    go install ./cmd/...

FROM debian:bullseye-20220822

LABEL org.opencontainers.image.source=https://github.com/simon-engledew/sqljson/

RUN --mount=type=cache,target=/var/cache/apt --mount=type=cache,target=/var/lib/apt \
    rm /etc/apt/apt.conf.d/docker-clean && apt-get update && apt-get install --yes --no-install-recommends graphviz=2.42.2-5

COPY --from=build /go/bin/sqljsondump /go/bin/sqljsondot /usr/bin/

USER nobody:nogroup

ENTRYPOINT ["/bin/sh", "-c", "sqljsondump | sqljsondot | dot -Tsvg"]

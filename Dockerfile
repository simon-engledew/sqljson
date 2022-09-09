FROM golang:1.19.1-bullseye AS build

WORKDIR /workspace

COPY . ./

RUN go install ./cmd/...

FROM debian:bullseye-20220822

LABEL org.opencontainers.image.source=https://github.com/simon-engledew/sqljson/

RUN apt-get update && apt-get install --yes --no-install-recommends graphviz

COPY --from=build /go/bin/sqljsondump /go/bin/sqljsondot /usr/bin/

ENTRYPOINT ["/bin/sh", "-c", "sqljsondump | sqljsondot | dot -Tsvg"]

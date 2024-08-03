# Build
FROM golang:alpine AS build

RUN apk add --no-cache -U build-base git make

RUN mkdir -p /src

WORKDIR /src

# Copy Makefile
COPY Makefile ./

# Copy go.mod and go.sum and install and cache dependencies
COPY go.mod .
COPY go.sum .

# Copy sources
COPY *.go ./
COPY ./internal/*.go ./internal/
COPY ./internal/codec/*.go ./internal/codec/
COPY ./internal/config/*.go ./internal/config/
COPY ./internal/data/*.go ./internal/data/
COPY ./internal/index/*.go ./internal/index/
COPY ./internal/metadata/*.go ./internal/metadata/
COPY ./cmd/bitcask/*.go ./cmd/bitcask/
COPY ./cmd/bitcaskd/*.go ./cmd/bitcaskd/

ARG VERSION="0.0.0"
ARG COMMIT="HEAD"
ARG BUILD=""

# Build binaries
RUN make build VERSION=$VERSION COMMIT=$COMMIT BUILD=$BUILD

# Runtime
FROM alpine:latest

RUN apk --no-cache -U add su-exec shadow ca-certificates tzdata ffmpeg

ENV PUID=1000
ENV PGID=1000

RUN addgroup -g "${PGID}" bitcaskd && \
    adduser -D -H -G bitcaskd -h /var/empty -u "${PUID}" bitcaskd && \
    mkdir -p /data && chown -R bitcaskd:bitcaskd /data

VOLUME /data

WORKDIR /

# force cgo resolver
ENV GODEBUG=netdns=cgo

COPY --from=build /src/bitcaskd /usr/local/bin/bitcaskd

COPY .dockerfiles/entrypoint.sh /init

ENTRYPOINT ["/init"]

CMD ["bitcaskd", "/data"]

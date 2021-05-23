ARG GOLANG_VERSION=1.16
FROM golang:${GOLANG_VERSION}-alpine
LABEL maintainer="Yuriy Gorbachev <yuriy@gorbachev.rocks>"

RUN apk update && apk upgrade
RUN apk add imagemagick libwebp libwebp-tools optipng

ARG MOZJPEG_VERSION=4.0.3
RUN set -x -o pipefail \
    && wget -O- https://github.com/mozilla/mozjpeg/archive/refs/tags/v${MOZJPEG_VERSION}.tar.gz | tar xzC /tmp \
    && apk add build-base make zlib zlib-static libpng libpng-static \
    && apk add --virtual mozjpeg-depencies nasm cmake zlib-dev libpng-dev \
    && cd /tmp/mozjpeg-${MOZJPEG_VERSION} \
    && mkdir build && cd build \
    && cmake -G"Unix Makefiles" ../ \
    && make install \
    && ln -s /opt/mozjpeg/bin/cjpeg /usr/bin/mozjpeg \
    && ln -s /opt/mozjpeg/bin/jpegtran /usr/bin/mozjpegtran \
    && rm -rf /tmp/mozjpeg-${MOZJPEG_VERSION} \
    && apk del --purge mozjpeg-depencies \
    && rm -rf /var/cache/apk/*

ARG ZOPFLI_VERSION=1.0.3
RUN set -x -o pipefail \
    && wget -O- https://github.com/google/zopfli/archive/zopfli-${ZOPFLI_VERSION}.tar.gz | tar xzC /tmp \
    && cd /tmp/zopfli-zopfli-${ZOPFLI_VERSION} \
    && make zopflipng \
    && cp ./zopflipng /usr/bin/ \
    && rm -rf /tmp/zopfli-zopfli-${ZOPFLI_VERSION}


ARG MODULE_PATH
ENV MODULE_ABS_PATH="/go/src/${MODULE_PATH}"

WORKDIR ${MODULE_ABS_PATH}
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENTRYPOINT go run cmd/gokaru/main.go
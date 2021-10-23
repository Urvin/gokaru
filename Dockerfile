ARG GOLANG_VERSION=1.17.1
FROM golang:${GOLANG_VERSION}-alpine
LABEL maintainer="Yuriy Gorbachev <yuriy@gorbachev.rocks>"

RUN apk update && apk upgrade
RUN apk add pngquant

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

ARG IMAGEMAGICK_VERSION=7.1.0-10
RUN set -x -o pipefail \
    && wget -O- https://github.com/ImageMagick/ImageMagick/archive/refs/tags/${IMAGEMAGICK_VERSION}.tar.gz | tar xzC /tmp \
    && apk add tiff fontconfig freetype libheif libwebp libxml2 \
    && apk add --virtual imagemagick-depencies zlib-dev libpng-dev libjpeg-turbo-dev freetype-dev fontconfig-dev libwebp-dev libtool tiff-dev lcms2-dev libheif-dev  libxml2-dev \
    && cd /tmp/ImageMagick-${IMAGEMAGICK_VERSION} \
    && ./configure --without-magick-plus-plus --without-perl --disable-openmp --with-gvc=no --disable-docs \
    && make -j$(nproc) \
    && make install \
    && ldconfig /usr/local/lib \
    && rm -rf /tmp/ImageMagick-${IMAGEMAGICK_VERSION} \
    && apk del --purge imagemagick-depencies \
    && rm -rf /var/cache/apk/*

ARG MODULE_PATH
ENV MODULE_ABS_PATH="/go/src/${MODULE_PATH}"

WORKDIR ${MODULE_ABS_PATH}
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENTRYPOINT go run cmd/gokaru/main.go
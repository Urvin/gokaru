ARG GOLANG_VERSION=1.17.2
ARG ALPINE_VERSION=3.14
FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION}
LABEL maintainer="Yuriy Gorbachev <yuriy@gorbachev.rocks>"

#----------------------------------------------------------------------------------------------------------------------#
# DEPENCIES VERSION
#----------------------------------------------------------------------------------------------------------------------#

ARG MOZJPEG_VERSION=4.0.3
ARG SPNG_VERSION=0.7.0
ARG TIFF_VERSION=4.3.0
ARG ZOPFLI_VERSION=1.0.3
ARG LCMS2_VERSION=2.12
ARG IMAGEMAGICK_VERSION=7.1.0-10
ARG VIPS_VERSION=8.11.4

ARG LIBDE265_VERSION=1.0.8
ARG DAV1D_VERSION=0.9.2
ARG LIBHEIF_VERSION=1.12.0
#----------------------------------------------------------------------------------------------------------------------#
# BUILD DEPENCIES
#----------------------------------------------------------------------------------------------------------------------#

RUN echo http://dl-cdn.alpinelinux.org/alpine/edge/testing >> /etc/apk/repositories

RUN apk update && apk upgrade
RUN apk add pkgconfig libtool git
RUN apk add libffi zlib zlib-static glib expat libxml2 libexif libpng libpng-static libwebp xz fftw libgsf orc giflib libimagequant rav1e
RUN apk add zlib-dev glib-dev expat-dev libxml2-dev libexif-dev libpng-dev libwebp-dev xz-dev fftw-dev libgsf-dev orc-dev giflib-dev libimagequant-dev rav1e-dev
RUN apk add build-base make nasm cmake meson curl
RUN go get -u -v github.com/ahmetb/govvv

ARG DEPS_PATH=/tmp/deps
RUN mkdir ${DEPS_PATH}

#MOZJPEG
RUN set -x -o pipefail \
    && wget -O- https://github.com/mozilla/mozjpeg/archive/refs/tags/v${MOZJPEG_VERSION}.tar.gz | tar xzC ${DEPS_PATH} \
    && cd ${DEPS_PATH}/mozjpeg-${MOZJPEG_VERSION} \
    && mkdir build && cd build \
    && cmake -DCMAKE_INSTALL_PREFIX=/usr/local -DCMAKE_INSTALL_DEFAULT_LIBDIR=lib .. \
    && make \
    && make install

#SPNG
RUN set -x -o pipefail \
    && wget -O- https://github.com/randy408/libspng/archive/refs/tags/v${SPNG_VERSION}.tar.gz | tar xzC ${DEPS_PATH} \
    && cd ${DEPS_PATH}/libspng-${SPNG_VERSION} \
    && meson build --buildtype=release \
    && cd build \
    && ninja \
    && ninja install

#TIFF
RUN set -x -o pipefail \
    && wget -O- https://gitlab.com/libtiff/libtiff/-/archive/v${TIFF_VERSION}/libtiff-v${TIFF_VERSION}.tar.gz | tar xzC ${DEPS_PATH} \
    && cd ${DEPS_PATH}/libtiff-v${TIFF_VERSION} \
    && cmake . \
    && make \
    && make install

#ZOPFLI
RUN set -x -o pipefail \
    && wget -O- https://github.com/google/zopfli/archive/zopfli-${ZOPFLI_VERSION}.tar.gz | tar xzC ${DEPS_PATH} \
    && cd ${DEPS_PATH}/zopfli-zopfli-${ZOPFLI_VERSION} \
    && make zopflipng \
    && cp ./zopflipng /usr/bin/

#LCMS
RUN set -x -o pipefail \
    && wget -O- https://sourceforge.net/projects/lcms/files/lcms/${LCMS2_VERSION}/lcms2-${LCMS2_VERSION}.tar.gz/download | tar xzC ${DEPS_PATH} \
    && cd ${DEPS_PATH}/lcms2-${LCMS2_VERSION} \
    && ./configure \
         --host=$HOST \
         --prefix=/usr/local \
         --enable-shared \
         --disable-static \
         --disable-dependency-tracking \
    && make install-strip

#LIBDE265
RUN set -x -o pipefail \
    && wget -O- https://github.com/strukturag/libde265/releases/download/v${LIBDE265_VERSION}/libde265-${LIBDE265_VERSION}.tar.gz | tar xzC ${DEPS_PATH} \
    && cd ${DEPS_PATH}/libde265-${LIBDE265_VERSION} \
    && ./configure --prefix=/usr/local --enable-shared --disable-static \
    && make install-strip

#DAV1D
RUN set -x -o pipefail \
    && wget -O- https://code.videolan.org/videolan/dav1d/-/archive/${DAV1D_VERSION}/dav1d-${DAV1D_VERSION}.tar.gz | tar xzC ${DEPS_PATH} \
    && cd ${DEPS_PATH}/dav1d-${DAV1D_VERSION} \
    && meson setup _build --buildtype=release --strip --prefix=/usr/local --libdir=lib \
    && ninja -C _build \
    && ninja -C _build install

#HEIF
RUN set -x -o pipefail \
    && wget -O- https://github.com/strukturag/libheif/releases/download/v${LIBHEIF_VERSION}/libheif-${LIBHEIF_VERSION}.tar.gz | tar xzC ${DEPS_PATH} \
    && cd ${DEPS_PATH}/libheif-${LIBHEIF_VERSION} \
    && mkdir _build \
    && cd _build \
    && cmake \
      -G"Unix Makefiles" \
      -DCMAKE_INSTALL_PREFIX=/usr/local \
      -DBUILD_SHARED_LIBS=1 \
      -DWITH_EXAMPLES=0 \
      .. \
    && make \
    && make install

#IMAGEMAGIC \
RUN set -x -o pipefail \
    && wget -O- https://github.com/ImageMagick/ImageMagick/archive/refs/tags/${IMAGEMAGICK_VERSION}.tar.gz | tar xzC ${DEPS_PATH} \
    && cd ${DEPS_PATH}/ImageMagick-${IMAGEMAGICK_VERSION} \
    && ./configure \
      --without-magick-plus-plus \
      --without-perl \
      --disable-openmp \
      --with-gvc=no \
      --disable-docs \
    && make -j$(nproc) \
    && make install \
    && ldconfig /usr/local/lib

RUN set -x -o pipefail \
    && wget -O- https://github.com/libvips/libvips/releases/download/v${VIPS_VERSION}/vips-${VIPS_VERSION}.tar.gz | tar xzC ${DEPS_PATH} \
    && cd ${DEPS_PATH}/vips-${VIPS_VERSION} \
    && CFLAGS=-O3 CXXFLAGS=-O3 ./configure \
        --enable-shared \
        --disable-static \
        --disable-dependency-tracking \
        --disable-debug \
        --disable-introspection \
        --without-python \
        --without-pangoft2 \
        --without-ppm  \
        --without-analyze \
        --without-radiance \
        --with-jpeg-includes=/usr/local/include \
        --with-jpeg-libraries=/usr/local/lib64 \
    && make V=0 \
    && make install

RUN rm -rf ${DEPS_PATH}
RUN rm -rf /var/cache/apk/*

#----------------------------------------------------------------------------------------------------------------------#

ARG MODULE_PATH
ENV MODULE_ABS_PATH="/go/src/${MODULE_PATH}"

RUN mkdir /var/log/gokaru/

WORKDIR ${MODULE_ABS_PATH}
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN govvv build -v -o /usr/local/bin/gokaru cmd/gokaru/main.go

CMD ["gokaru"]
EXPOSE 8081

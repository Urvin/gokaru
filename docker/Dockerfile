ARG BUILDER_VERSION=1.0.2
ARG ALPINE_VERSION=3.21
FROM urvinio/gokaru-builder:${BUILDER_VERSION} AS builder
LABEL maintainer="Yuriy Gorbachev <yuriy@gorbachev.rocks>"

ARG MODULE_PATH
ENV MODULE_ABS_PATH="/go/src/${MODULE_PATH}"

WORKDIR ${MODULE_ABS_PATH}

COPY ../VERSION .
RUN VERSION=$(cat VERSION)

ENV CGO_ENABLED=1

COPY ../go.mod go.sum ./
RUN go mod download
COPY .. .
RUN go build -ldflags="-X github.com/urvin/gokaru/internal/version.Version=${VERSION}" -v -o /usr/local/bin/gokaru cmd/gokaru/main.go

RUN apk del --purge build-depencies
RUN rm -rf ${DEPS_PATH}
RUN rm -rf /var/cache/apk/*

#----------------------------------------------------------------------------------------------------------------------#
# RUNTIME
#----------------------------------------------------------------------------------------------------------------------#
FROM alpine:${ALPINE_VERSION}
LABEL maintainer="Yuriy Gorbachev <yuriy@gorbachev.rocks>"

RUN mkdir -p /var/gokaru/storage/
RUN mkdir -p /var/gokaru/assets/
RUN mkdir -p /var/gokaru/config/

RUN echo http://dl-cdn.alpinelinux.org/alpine/edge/testing >> /etc/apk/repositories
RUN apk update && apk upgrade
RUN apk add libffi zlib zlib-static glib expat libxml2 libexif libpng libpng-static libwebp xz fftw libgsf orc giflib libimagequant rav1e rav1e-dev libstdc++
RUN apk add libwebpmux libwebpdemux libhwy

ARG MODULE_PATH

COPY --from=builder /usr/local/bin/gokaru /usr/local/bin/
COPY --from=builder /usr/local/bin/convert /usr/local/bin
COPY --from=builder /usr/bin/zopflipng /usr/bin
COPY --from=builder /usr/local/lib /usr/local/lib

COPY --from=builder /go/src/${MODULE_PATH}/config/config.yml /var/gokaru/config/
COPY --from=builder /go/src/${MODULE_PATH}/assets/favicon.ico /var/gokaru/assets/
COPY --from=builder /go/src/${MODULE_PATH}/assets/error.html /var/gokaru/assets/

CMD ["gokaru"]
EXPOSE 80

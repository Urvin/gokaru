ARG BUILDER_VERSION=1.0.2
ARG ALPINE_VERSION=3.20
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
#RUN go build -ldflags="-X github.com/urvin/gokaru/internal/version.Version=dev" -v -o /usr/local/bin/gokaru cmd/gokaru/main.go

RUN mkdir -p /var/gokaru/storage/
RUN mkdir -p /var/gokaru/assets/
RUN mkdir -p /var/gokaru/config/

CMD tail -f /dev/null

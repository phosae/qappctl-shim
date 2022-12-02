FROM --platform=$BUILDPLATFORM golang:1.18 as builder
ARG TARGETOS TARGETARCH
WORKDIR /workspace
ENV GOPROXY=https://goproxy.cn,direct
COPY . /
RUN go mod download
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -o qappctl-shim-$TARGETARCH main.go

FROM alpine:3.17
RUN set -ex \
    && apk update \
    && apk upgrade \
    && apk add --no-cache \
    bash \
    docker-cli
ARG TARGETARCH
WORKDIR /
ADD /ctl/qappctl-linux-$TARGETARCH /usr/bin/
COPY --from=builder /workspace/qappctl-shim-$TARGETARCH /qappctl-shim

ENTRYPOINT ["/qappctl-shim"]

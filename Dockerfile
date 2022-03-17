FROM golang:1.18-alpine AS builder
WORKDIR /go/src/go.lstv.dev/goproxy/
ARG GOPROXY_VERSION=unknown
COPY . .
RUN go build \
    -ldflags "-X go.lstv.dev/goproxy/util.version=$GOPROXY_VERSION" \
    -o goproxy \
    cmd/goproxy/main.go
RUN chmod +x ./goproxy

FROM alpine:latest
ARG BUILD_DATE
LABEL io.k8s.display-name="Livesport TV: Go Proxy"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.authors="goproxy@livesporttv.cz"
LABEL org.opencontainers.image.url="https://github.com/livesport-tv/goproxy/"
LABEL org.opencontainers.image.documentation="https://github.com/livesport-tv/goproxy/"
LABEL org.opencontainers.image.source="https://github.com/livesport-tv/goproxy/"
LABEL org.opencontainers.image.version="${GOPROXY_VERSION}"
LABEL org.opencontainers.image.vendor="Livesport TV"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.title="Go Proxy"
LABEL org.opencontainers.image.description="Livesport TV Go Proxy with monorepo support."
RUN mkdir -p /etc/lstv/goproxy
COPY --from=builder /go/src/go.lstv.dev/goproxy/goproxy /usr/local/bin/
CMD ["/etc/lstv/goproxy/config.json"]
ENTRYPOINT ["goproxy"]

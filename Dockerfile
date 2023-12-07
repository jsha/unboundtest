FROM alpine:latest AS unboundtest
RUN apk update
RUN apk add go
COPY *.go go.* /unboundtest-repo/
WORKDIR /unboundtest-repo
RUN GOBIN=/usr/bin CGO_ENABLED=0 go install .

FROM alpine:latest AS unbound
ARG UNBOUND_VERSION=1.18.0
RUN apk update
RUN apk add curl
RUN curl -o unbound.tgz https://nlnetlabs.nl/downloads/unbound/unbound-$UNBOUND_VERSION.tar.gz
RUN tar xzf unbound.tgz
RUN apk add \
  flex \
  bison \
  openssl-dev \
  openssl-libs-static \
  libexpat \
  expat-static \
  expat-dev \
  libev-dev \
  build-base
WORKDIR unbound-$UNBOUND_VERSION
RUN ./configure --enable-fully-static && make
RUN install unbound /usr/sbin/unbound

FROM gcr.io/distroless/base-debian12
LABEL org.opencontainers.image.source=https://github.com/jsha/unboundtest
COPY --from=unboundtest /usr/bin/unboundtest /usr/bin/unboundtest
COPY --from=unbound /usr/sbin/unbound /usr/sbin/unbound
COPY index.html root.key /work/
COPY unbound.conf /etc/unbound/
WORKDIR /work/
EXPOSE 1232
CMD ["/usr/bin/unboundtest", "-config", "/etc/unbound/unbound.conf"]

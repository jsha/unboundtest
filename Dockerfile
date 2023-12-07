ARG UNBOUND_VERSION=1.18
FROM alpine:latest AS unboundtest
RUN apk update
RUN apk add go
COPY *.go go.* /unboundtest-repo/
WORKDIR /unboundtest-repo
RUN GOBIN=/usr/bin CGO_ENABLED=0 go install .

FROM alpine:latest AS unbound
RUN apk update
RUN apk add \
  ca-certificates \
  git \
  flex \
  bison \
  openssl-dev \
  openssl-libs-static \
  libexpat \
  expat-static \
  expat-dev \
  libev-dev \
  build-base
RUN git clone --depth 1 -b release-$UNBOUND_VERSION https://github.com/NLnetLabs/unbound/ /unbound
WORKDIR /unbound
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

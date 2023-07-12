FROM ubuntu:lunar
LABEL org.opencontainers.image.source=https://github.com/jsha/unboundtest
ENV DEBIAN_FRONTEND noninteractive
RUN apt update && apt -y install unbound golang-go ca-certificates
COPY . /unboundtest
WORKDIR /unboundtest
RUN GOBIN=/usr/bin go install ./
RUN mkdir -p /var/run/unboundtest
COPY index.html root.key unbound.conf /var/run/unboundtest/
WORKDIR /var/run/unboundtest
EXPOSE 1232
CMD ["/usr/bin/unboundtest"]

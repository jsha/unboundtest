FROM ubuntu:23.04
LABEL org.opencontainers.image.source=https://github.com/jsha/unboundtest
ENV DEBIAN_FRONTEND noninteractive
RUN apt update && apt -y install unbound golang-go ca-certificates
COPY . /unboundtest-repo
WORKDIR /unboundtest-repo
RUN GOBIN=/usr/bin go install .
RUN mkdir -p /work/
COPY index.html root.key unbound.conf /work/
WORKDIR /work/
EXPOSE 1232
CMD ["/usr/bin/unboundtest"]

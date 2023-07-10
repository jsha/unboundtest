FROM ubuntu:23.04
ENV DEBIAN_FRONTEND noninteractive
RUN apt update && apt -y install unbound golang-go ca-certificates git
COPY . /unboundtest
WORKDIR /unboundtest
RUN GOBIN=/usr/bin go install ./
RUN mkdir -p /var/run/unboundtest
COPY index.html root.key unbound.conf /var/run/unboundtest/
COPY root.key unbound.conf /etc/unbound
WORKDIR /var/run/unboundtest
EXPOSE 1232
CMD ["/usr/bin/unboundtest"]

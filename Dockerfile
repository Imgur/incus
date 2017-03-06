FROM golang:1.8
RUN mkdir /etc/incus

ADD . /go/src/github.com/jtaylor32/incus
WORKDIR /go/src/github.com/jtaylor32/incus

RUN ./scripts/build.sh
RUN mkdir -p /etc/incus

CMD /go/bin/incus -conf="/etc/incus/"

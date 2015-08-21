FROM golang:1.4

RUN mkdir /etc/incus

ADD . /go/src/github.com/Imgur/incus
WORKDIR /go/src/github.com/Imgur/incus

ADD incus.conf /etc/incus/incus.conf

RUN ./scripts/build.sh

CMD ["/go/bin/incus"]

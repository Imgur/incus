FROM golang:1.8
RUN mkdir /etc/incus

ADD . /go/src/github.com/Imgur/incus
WORKDIR /go/src/github.com/Imgur/incus

RUN ./scripts/build.sh
RUN mkdir -p /etc/incus

CMD /go/bin/incus -conf="/etc/incus/"

FROM golang:1.4
ADD ./scripts/build-setup.sh /tmp/build-setup.sh
RUN /tmp/build-setup.sh

ADD . /go/src/github.com/Imgur/incus
WORKDIR /go/src/github.com/Imgur/incus

RUN ./scripts/build-build.sh

CMD incus

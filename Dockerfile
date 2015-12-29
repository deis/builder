FROM golang:1.5

ENV VERSION 0.0.1
ENV GO15VENDOREXPERIMENT 1
ENV CGO_ENABLED 0
ENV LDFLAGS "-s -X main.version=$VERSION"
ENV BINDIR rootfs/bin
ENV DEIS_RELEASE=2.0.0-dev

WORKDIR /app
RUN go build -o $BINDIR/boot -a -installsuffix cgo -ldflags "$LDFLAGS"  boot.go

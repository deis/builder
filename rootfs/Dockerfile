FROM alpine:3.1

# install common packages
RUN apk add --update-cache curl bash sudo && rm -rf /var/cache/apk/*

# install etcdctl
RUN curl -sSL -o /usr/local/bin/etcdctl https://s3-us-west-2.amazonaws.com/get-deis/etcdctl-v0.4.9 \
    && chmod +x /usr/local/bin/etcdctl

# install confd
RUN curl -sSL -o /usr/local/bin/confd https://github.com/kelseyhightower/confd/releases/download/v0.10.0/confd-0.10.0-linux-amd64 \
    && chmod +x /usr/local/bin/confd

RUN apk add --update-cache \
    coreutils \
    device-mapper \
    e2fsprogs \
    git \
    iptables \
    libudev \
    lxc \
    openssh \
    udev \
    util-linux \
    xz \
    && rm -rf /var/cache/apk/*


# configure ssh server
RUN mkdir -p /var/run/sshd && rm -rf /etc/ssh/ssh_host*

# install git and configure gituser
ENV GITHOME /home/git
ENV GITUSER git
# this is so the minio client (https://github.com/minio/mc) works properly
ENV DOCKERIMAGE=1
RUN mkdir /apps
RUN adduser -D -h $GITHOME $GITUSER
RUN mkdir -p $GITHOME/.ssh && chown git:git $GITHOME/.ssh
RUN chown -R $GITUSER:$GITUSER $GITHOME


ADD ./check-repos /home/git/
ENTRYPOINT ["/bin/boot", "server"]
EXPOSE 2223
EXPOSE 3000
RUN addgroup -g 2000 slug && adduser -D -u 2000 -G slug slug

# $GITUSER is added to docker group to use docker without sudo and to slug
# group in order to share resources with the slug user
RUN addgroup $GITUSER slug
RUN passwd -u git

COPY . /

ENV DEIS_RELEASE 2.0.0-dev

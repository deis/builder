FROM alpine:3.3

# install git and configure gituser
ENV GITHOME /home/git
ENV GITUSER git
# this is so the minio client (https://github.com/minio/mc) works properly
ENV DOCKERIMAGE=1
ENV DEIS_RELEASE 2.0.0-dev
ENV DEIS_BUILDER_SERVER_FETCHER_PORT=3000
ENV DEIS_BUILDER_SERVER_SSH_HOST_IP=0.0.0.0
ENV DEIS_BUILDER_SERVER_SSH_HOST_PORT=2223


# install common packages
RUN apk add --update-cache \
  curl \
  bash \
  sudo \
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

# install etcdctl & confd, configure the SSH server, then set up directories, users and perms.
# NOTE: $GITUSER is added to docker group to use docker without sudo and to slug group in order to share resources with the slug user
RUN curl -sSL -o /usr/local/bin/etcdctl https://s3-us-west-2.amazonaws.com/get-deis/etcdctl-v0.4.9 \
    && chmod +x /usr/local/bin/etcdctl \
    && curl -sSL -o /usr/local/bin/confd https://github.com/kelseyhightower/confd/releases/download/v0.10.0/confd-0.10.0-linux-amd64 \
    && chmod +x /usr/local/bin/confd \
    && mkdir -p /var/run/sshd && rm -rf /etc/ssh/ssh_host* \
    && mkdir /apps \
    && adduser -D -h $GITHOME $GITUSER \
    && mkdir -p $GITHOME/.ssh && chown git:git $GITHOME/.ssh \
    && chown -R $GITUSER:$GITUSER $GITHOME \
    && addgroup -g 2000 slug && adduser -D -u 2000 -G slug slug \
    && addgroup $GITUSER slug \
    && passwd -u git


ADD ./check-repos /home/git/
ENTRYPOINT ["boot", "server"]
EXPOSE 2223
EXPOSE 3000

COPY . /

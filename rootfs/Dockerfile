FROM alpine:3.3

# install git and configure gituser
ENV GITHOME /home/git
ENV GITUSER git

# install common packages
RUN apk add --update-cache \
  bash \
  sudo \
  git \
  openssh \
  coreutils \
  tar \
  xz \
  && rm -rf /var/cache/apk/*

# install & configure the SSH server, then set up directories, users and perms.
# NOTE: $GITUSER is added to docker group to use docker without sudo and to slug group in order to share resources with the slug user
RUN mkdir -p /var/run/sshd && rm -rf /etc/ssh/ssh_host* \
    && mkdir /apps \
    && adduser -D -h $GITHOME $GITUSER \
    && mkdir -p $GITHOME/.ssh && chown git:git $GITHOME/.ssh \
    && chown -R $GITUSER:$GITUSER $GITHOME \
    && addgroup -g 2000 slug && adduser -D -u 2000 -G slug slug \
    && addgroup $GITUSER slug \
    && passwd -u git

COPY . /

# this is so the minio client (https://github.com/minio/mc) works properly
ENV DOCKERIMAGE=1
ENV DEIS_RELEASE 2.0.0-dev
ENV DEIS_BUILDER_SERVER_FETCHER_PORT=3000
ENV DEIS_BUILDER_SERVER_SSH_HOST_IP=0.0.0.0
ENV DEIS_BUILDER_SERVER_SSH_HOST_PORT=2223

ENTRYPOINT ["boot", "server"]
EXPOSE 2223
EXPOSE 3000

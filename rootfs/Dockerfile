FROM alpine:3.3

# specify the git user's home and name
ENV GITHOME /home/git
ENV GITUSER git

ENV DEIS_BUILDER_SERVER_SSH_HOST_IP=0.0.0.0
ENV DEIS_BUILDER_SERVER_SSH_HOST_PORT=2223

EXPOSE 2223
EXPOSE 3000

# install common packages, install & configure the SSH server, and set up directories, users & perms.
# NOTE: $GITUSER is added to docker group to use docker without sudo and to slug group in order to share resources with the slug user
RUN apk add --update-cache \
  bash \
  sudo \
  git \
  openssh \
  coreutils \
  tar \
  xz \
  && rm -rf /var/cache/apk/* \
  && mkdir -p /var/run/sshd && rm -rf /etc/ssh/ssh_host* \
  && mkdir /apps \
  && adduser -D -h $GITHOME $GITUSER \
  && mkdir -p $GITHOME/.ssh && chown git:git $GITHOME/.ssh \
  && chown -R $GITUSER:$GITUSER $GITHOME \
  && addgroup -g 2000 slug && adduser -D -u 2000 -G slug slug \
  && addgroup $GITUSER slug \
  && passwd -u git

COPY . /

ENV WORKFLOW_RELEASE 2.0.0-dev

ENTRYPOINT ["boot", "server"]

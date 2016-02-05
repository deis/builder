FROM alpine:3.1

# install common packages
RUN apk add --update-cache \
  curl \
  bash \
  sudo \
  git \
  openssh \
  coreutils \
  tar \
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

ENTRYPOINT ["boot", "server"]
EXPOSE 2223
EXPOSE 3000

RUN addgroup -g 2000 slug && adduser -D -u 2000 -G slug slug
RUN addgroup $GITUSER slug
RUN passwd -u git

COPY . /

ENV DEIS_RELEASE 2.0.0-dev
ENV DEIS_BUILDER_SERVER_FETCHER_PORT=3000
ENV DEIS_BUILDER_SERVER_SSH_HOST_IP=0.0.0.0
ENV DEIS_BUILDER_SERVER_SSH_HOST_PORT=2223

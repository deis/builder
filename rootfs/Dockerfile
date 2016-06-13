FROM quay.io/deis/base:0.2.0

RUN adduser --system \
	--shell /bin/bash \
	--disabled-password \
	--home /home/git \
	--group \
	git

COPY . /

RUN apt-get update \
	&& apt-get install -y \
		sudo \
		openssh-server \
		coreutils \
		tar \
		xz-utils \
		--no-install-recommends \
	&& mkdir -p /var/run/sshd \
	&& rm -rf /etc/ssh/ssh_host* \
	&& mkdir /apps \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* /usr/share/man /usr/share/doc

# install git from source
ENV GIT_VERSION 2.8.4
RUN apt-get update \
	&& apt-get install -y \
		build-essential \
		gettext \
		libcurl4-gnutls-dev \
		libexpat1-dev \
		libssl-dev \
		libz-dev \
		--no-install-recommends \
	&& curl -SsL https://github.com/git/git/archive/v$GIT_VERSION.tar.gz --output git.tar.gz \
	&& tar -xzf git.tar.gz \
	&& cd git-$GIT_VERSION \
	&& make prefix=/usr/local all \
	&& make prefix=/usr/local install \
	&& cd .. \
	&& rm -rf git-$GIT_VERSION git.tar.gz \
	&& passwd -u git \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* /usr/share/man /usr/share/doc

CMD ["/usr/bin/boot", "server"]
EXPOSE 2223
EXPOSE 3000

ENV WORKFLOW_RELEASE 2.0.0

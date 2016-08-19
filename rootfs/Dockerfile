FROM quay.io/deis/base:0.3.1

RUN adduser --system \
	--shell /bin/bash \
	--disabled-password \
	--home /home/git \
	--group \
	git

RUN apt-key adv --keyserver keyserver.ubuntu.com --recv-keys E1DF1F24 \
	&& echo "deb http://ppa.launchpad.net/git-core/ppa/ubuntu xenial main" >> /etc/apt/sources.list \
	&& apt-get update \
	&& apt-get install -y \
		git \
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
	&& rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* /usr/share/man /usr/share/doc \
	&& passwd -u git

COPY . /

CMD ["/usr/bin/boot", "server"]
EXPOSE 2223
EXPOSE 3000

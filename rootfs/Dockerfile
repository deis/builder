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
		git \
		openssh-server \
		coreutils \
		tar \
		xz-utils \
		--no-install-recommends \
	&& mkdir -p /var/run/sshd \
	&& rm -rf /etc/ssh/ssh_host* \
	&& mkdir /apps \
	&& passwd -u git \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* /usr/share/man /usr/share/doc

CMD ["/usr/bin/boot", "server"]
EXPOSE 2223
EXPOSE 3000

ENV WORKFLOW_RELEASE 2.0.0

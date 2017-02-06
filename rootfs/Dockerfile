FROM quay.io/deis/base:v0.3.6

RUN adduser --system \
	--shell /bin/bash \
	--disabled-password \
	--home /home/git \
	--group \
	git

RUN apt-key adv --keyserver keyserver.ubuntu.com --recv-keys E1DF1F24 && \
	echo "deb http://ppa.launchpad.net/git-core/ppa/ubuntu xenial main" >> /etc/apt/sources.list && \
	apt-get update && \
	apt-get install -y --no-install-recommends \
		git \
		sudo \
		openssh-server \
		coreutils \
		tar \
		xz-utils && \
	mkdir -p /var/run/sshd && \
	rm -rf /etc/ssh/ssh_host* && \
	mkdir /apps && \
	passwd -u git && \
    # cleanup
    apt-get autoremove -y && \
    apt-get clean -y && \
    # package up license files if any by appending to existing tar
    COPYRIGHT_TAR='/usr/share/copyrights.tar'; \
    gunzip -f $COPYRIGHT_TAR.gz; tar -rf $COPYRIGHT_TAR /usr/share/doc/*/copyright; gzip $COPYRIGHT_TAR && \
    rm -rf \
        /usr/share/doc \
        /usr/share/man \
        /usr/share/info \
        /usr/share/locale \
        /var/lib/apt/lists/* \
        /var/log/* \
        /var/cache/debconf/* \
        /etc/systemd \
        /lib/lsb \
        /lib/udev \
        /usr/lib/x86_64-linux-gnu/gconv/IBM* \
        /usr/lib/x86_64-linux-gnu/gconv/EBC* && \
    bash -c "mkdir -p /usr/share/man/man{1..8}"

COPY . /

CMD ["/usr/bin/boot", "server"]
EXPOSE 2223
EXPOSE 3000

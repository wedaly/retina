# mcr.microsoft.com/azurelinux/base/core:3.0.20241101
FROM mcr.microsoft.com/azurelinux/base/core@sha256:87f833180f7ff6987231c1946995cce01d19ffbc7f815b6444b714cec63c8310

# Known issue in Azure Linux 3:
# `nslookup` and `bind` print an error "Algorithm not supported by SCOSSL" (but still complete successfully).
# This will be fixed by https://github.com/microsoft/SymCrypt-OpenSSL/pull/92
RUN tdnf install -y \
	bind-utils \
	conntrack \
	curl \
	ebtables-legacy \
	ethtool \
	iperf3 \
	iproute \
	ipset \
	iptables \
	iputils \
	jq \
	ldns-utils \
	net-tools \
	nftables \
	nmap \
	nmap-ncat \
	openssh \
	socat \
	tcpdump \
	wget \
	&& tdnf clean all

CMD ["/bin/bash"]

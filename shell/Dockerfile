FROM mcr.microsoft.com/azurelinux/base/core:3.0

# TODO: fix this error from nslookup and dig
# [ERROR] error:41000066:SCOSSL::Algorithm not supported by SCOSSL:SCOSSL does not support hash algorithm for MAC 4 at /usr/src/azl/BUILD/SymCrypt-OpenSSL-1.5.1/ScosslCommon/src/scossl_mac.c, line 193
RUN tdnf install -y \
	bind-utils \
	bpftool \
	conntrack \
	curl \
	ebtables-legacy \
	iperf3 \
	iproute \
	ipset \
	iptables \
	iputils \
	ldns-utils \
	net-tools \
	nftables \
	nmap \
	openssh \
	socat \
	tcpdump \
	wget \
	&& tdnf clean all

# TODO: set alternatives iptables -> iptables-nft

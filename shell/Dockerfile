FROM mcr.microsoft.com/azurelinux/base/core:3.0 # TODO: pin

# TODO: fix this error
# [ERROR] error:41000066:SCOSSL::Algorithm not supported by SCOSSL:SCOSSL does not support hash algorithm for MAC 4 at /usr/src/azl/BUILD/SymCrypt-OpenSSL-1.5.1/ScosslCommon/src/scossl_mac.c, line 193
RUN tdnf install -y \
	curl \
	iptables \
	conntrack \
	tcpdump \
	bind-utils \
	iproute \
	wget \
	iputils \
	ldns-utils \
	nftables \
	ebtables-legacy \
	bpftool \
	ipset \
	nmap \
	socat \
	openssh \
	iperf3 \
	net-tools

# TODO: set alternatives iptables -> iptables-nft

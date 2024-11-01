FROM mcr.microsoft.com/azurelinux/base/core:3.0

# There are a two known issues with Azure Linux 3.0.20241005 that affect this image:
# 1. `iptables-nft` binary is not yet installed, but will be fixed by https://github.com/microsoft/azurelinux/pull/10786
#    Until then, use `nft` to view nftables rules.
# 2. `nslookup` and `bind` print an error "Algorithm not supported by SCOSSL" (but still complete successfully).
#    This will be fixed by https://github.com/microsoft/SymCrypt-OpenSSL/pull/92
RUN tdnf install -y \
	bind-utils \
	bpftool \
	bpftrace \
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

CMD ["/bin/bash"]

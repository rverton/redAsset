package main

import (
	"net"
	"strings"
	"testing"
)

func TestParseDomains(t *testing.T) {
	data := "robinverton.de\ngoogle.de"

	domains, _, err := parseDomainsOrCIDR(strings.NewReader(data))
	if err != nil {
		t.Error(err)
	}

	if len(domains) != 2 {
		t.Error("did not parse 2 domains")
	}
	return
}

func TestParseDomainsCIDR(t *testing.T) {
	data := "robinverton.de\ngoogle.de\n192.168.1.1/24\n10.0.0.1/32"

	_, ips, err := parseDomainsOrCIDR(strings.NewReader(data))
	if err != nil {
		t.Error(err)
	}

	if len(ips) != 2 {
		t.Error("did not parse 2 domains")
	}

	if !ips[0].Contains(net.ParseIP("192.168.1.2")) {
		t.Error("IPs did not contain")
	}
	return
}

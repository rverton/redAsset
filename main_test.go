package main

import (
	"net"
	"testing"
)

var d = []DNSEntry{
	DNSEntry{
		Name:  "a.foobar.com",
		Value: "192.168.178.1",
	},
	DNSEntry{
		Name:  "b.foobar2.com",
		Value: "192.168.1.2",
	},
}

func TestIsValidResultIP(t *testing.T) {
	var ips []*net.IPNet
	_, net1, _ := net.ParseCIDR("192.168.1.1/24")
	_, net2, _ := net.ParseCIDR("10.0.0.1/24")
	ips = append(ips, net1)
	ips = append(ips, net2)

	if isValidResult(d[0], []string{}, []string{}, ips) {
		t.Errorf("ip %v wrongly declared as valid", d[0])
	}

	if !isValidResult(d[1], []string{}, []string{}, ips) {
		t.Errorf("could not find %v in ip list", d[1])
	}

}

func TestIsValidResultDomain(t *testing.T) {
	var ips []*net.IPNet

	allowed := []string{"foobar.com"}
	blacklist := []string{"a.foobar.com"}

	// test a valid result
	if !isValidResult(d[0], allowed, []string{}, ips) {
		t.Errorf("%v should be a result", d[0])
	}

	// test an invalid result
	if isValidResult(d[1], allowed, []string{}, ips) {
		t.Errorf("%v should not be a valid result", d[1])
	}

	// test blacklist
	if isValidResult(d[1], allowed, blacklist, ips) {
		t.Errorf("%v should not be a valid result", d[1])
	}
}

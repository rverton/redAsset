package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

type DNSEntry struct {
	Timestamp string
	Name      string
	Type      string
	Value     string
}

func readLine(r *bufio.Reader) ([]byte, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return ln, err
}

func parseDnsHosts(filename string) <-chan DNSEntry {

	ch := make(chan DNSEntry)

	go func() {
		var f io.ReadCloser

		file, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		// on-the-fly gzip
		if strings.HasSuffix(filename, "gz") {
			var err error
			if f, err = gzip.NewReader(file); err != nil {
				log.Fatalf("Error using gzip on file: %v", err)
			}

		} else {
			f = file
		}

		var e DNSEntry

		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
			t := scanner.Bytes()

			err := json.Unmarshal(t, &e)
			if err != nil {
				log.Printf("Error while unmarshaling: %v", err)
				continue
			}

			ch <- e

		}
		close(ch)
	}()

	return ch

}

func parseDomainFile(filename string) ([]string, []*net.IPNet, error) {

	handle, err := os.Open(filename)
	if err != nil {
		return []string{}, []*net.IPNet{}, err
	}
	defer handle.Close()

	return parseDomainsOrCIDR(handle)

}

func parseDomainsOrCIDR(handle io.Reader) ([]string, []*net.IPNet, error) {

	var domains []string
	var ips []*net.IPNet

	scanner := bufio.NewScanner(handle)

	for scanner.Scan() {
		t := scanner.Text()

		// parse domain or CIDR IPNet
		_, ipv4Net, err := net.ParseCIDR(t)
		if err != nil {
			s := fmt.Sprintf(".%v", t)
			domains = append(domains, s)
		} else {
			ips = append(ips, ipv4Net)
		}

	}

	return domains, ips, nil
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var jsonOutput *json.Encoder
var count uint64
var valid uint64
var workers int
var tt = time.Now()

func main() {
	parseFile := flag.String("file", "", "JSON file to parse from, gzip allowed.")
	parseDomainFilter := flag.String("domains", "", "File containing 2nd level domains to include.")
	parseDomainBlacklist := flag.String("bdomains", "", "File containing 2nd level domains to exclude.")
	workers = *flag.Int("workers", 4, "Number of workers to start.")
	useCATrans := flag.Bool("catransoff", false, "Deactivate querying certificate transparency logs (crt.sh).")

	flag.Parse()

	var allowedDomains []string
	var blacklistDomains []string
	var ips []*net.IPNet
	var wg sync.WaitGroup

	if *parseFile == "" {
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *parseDomainFilter != "" {
		var err error
		allowedDomains, ips, err = parseDomainFile(*parseDomainFilter)
		if err != nil {
			log.Fatalf("Error reading domain file: %v", err)
		}

		log.Printf("Limiting to %v domains and %v CIDRs.", len(allowedDomains), len(ips))
	}

	if *parseDomainBlacklist != "" {
		var err error
		blacklistDomains, ips, err = parseDomainFile(*parseDomainBlacklist)
		if err != nil {
			log.Fatalf("Error reading blacklist domain file: %v", err)
		}

		log.Printf("Limiting to %v blacklisted 2nd-lvl domains.", len(blacklistDomains))
	}

	wg.Add(1)
	go func() {
		log.Println("Starting FDNS search.")
		parseFDNS(*parseFile, allowedDomains, blacklistDomains, ips, &wg)
	}()

	if !*useCATrans {
		wg.Add(1)
		go func() {
			log.Println("Querying certificate transparency logs.")
			queryCATransparency(allowedDomains, blacklistDomains)
			wg.Done()
		}()
	}

	wg.Wait()

	log.Println("Finished.")
}

func queryCATransparency(allowed []string, blacklist []string) {
	var bodyDomain []struct {
		Domain string `json:"name_value"`
	}

	for _, domain := range allowed {
		url := fmt.Sprintf("https://crt.sh/?q=%%%v&output=json", domain)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error contacting crt.sh: %s", err)
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		json.Unmarshal([]byte(body), &bodyDomain)

		for _, d := range bodyDomain {
			if isValidResult(DNSEntry{Name: d.Domain}, allowed, blacklist, []*net.IPNet{}) {
				fmt.Println(d.Domain)
				atomic.AddUint64(&valid, 1)
			}
		}

		log.Printf("CA transparency: Got %v certificates for '%v'", len(bodyDomain), domain)
	}
}

func parseFDNS(fname string, allowed []string, blacklist []string, ips []*net.IPNet, wg *sync.WaitGroup) {

	if len(allowed) <= 0 && len(ips) <= 0 {
		log.Fatal("No valid domains (0) and IPs (0) parsed from input.")
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			for dnsentry := range parseDnsHosts(fname) {

				c := atomic.AddUint64(&count, 1)

				if isValidResult(dnsentry, allowed, blacklist, ips) {
					fmt.Println(dnsentry.Name)
					atomic.AddUint64(&valid, 1)
				}

				if c%5000000 == 0 && c > 0 {
					log.Printf("FDNS: %vm processed, %v valid (took %v)", c/1000000, atomic.LoadUint64(&valid), time.Since(tt))
					tt = time.Now()
				}
			}
			wg.Done()
		}()
	}
}

func isValidResult(dnsentry DNSEntry, allowed []string, blacklist []string, ips []*net.IPNet) bool {

	//check if IP is in one of the parsed networks
	if len(ips) > 0 {
		entryIp := net.ParseIP(dnsentry.Value)
		for _, ip := range ips {
			if ip.Contains(entryIp) {
				log.Printf("IP matched: %v (%v) in %v", dnsentry.Name, entryIp, ip)
				return true
			}
		}

		// if no allowed domains are passed, stop here
		if len(allowed) <= 0 {
			return false
		}
	}

	// check if allowed domain
	if len(allowed) > 0 {
		if !isAllowed(allowed, dnsentry.Name) {
			return false
		}
	}

	// remove blacklisted domains
	if len(blacklist) > 0 {
		if isAllowed(blacklist, dnsentry.Name) {
			return false
		}
	}

	return true
}

func isAllowed(allowed []string, domain string) bool {

	for _, s := range allowed {
		if strings.HasSuffix(domain, s) {
			return true
		}
	}
	return false
}

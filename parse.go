package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

func parseDnsHosts(filename string) <-chan string {

	ch := make(chan string)

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

		r := bufio.NewReader(f)

		var line []byte
		var e DNSEntry

		for {
			line, err = readLine(r)
			if err != nil {
				break
			}

			err := json.Unmarshal(line, &e)
			if err != nil {
				log.Printf("Error while unmarshaling: %v", err)
				continue
			}

			ch <- e.Name

		}
		close(ch)
	}()

	return ch

}

func parseDomainFile(filename string) ([]string, error) {

	var domains []string

	file, err := os.Open(filename)
	if err != nil {
		return domains, err
	}
	defer file.Close()

	r := bufio.NewReader(file)
	var line []byte

	for {
		line, err = readLine(r)
		if err != nil {
			break
		}

		s := fmt.Sprintf(".%v", string(line))

		domains = append(domains, s)
	}

	return domains, nil
}

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

type Host struct {
	Ip   string
	Data struct {
		Http struct {
			Response Response
		}
	}
}

type Response struct {
	Headers map[string][]string
	Body    string
}

type DNSEntry struct {
	Timestamp string
	Name      string
	Type      string
	Value     string
}

func (r *Response) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Headers map[string]json.RawMessage `json:"headers"`
		Body    string                     `json:"body"`
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	headers := make(map[string][]string)
	for k, v := range tmp.Headers {
		if k == "unknown" {
			var unknown []struct {
				Key   string
				Value []string
			}
			if err := json.Unmarshal(v, &unknown); err != nil {
				return err
			}
			for _, u := range unknown {
				headers[u.Key] = u.Value
			}
		} else {
			var values []string
			if err := json.Unmarshal(v, &values); err != nil {
				return err
			}
			headers[k] = values
		}
	}
	*r = Response{
		Headers: headers,
		Body:    tmp.Body,
	}
	return nil
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

func parseHttpHosts(filename string) <-chan Host {

	ch := make(chan Host)

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
		var host Host

		for {
			line, err = readLine(r)
			if err != nil {
				break
			}

			err := json.Unmarshal(line, &host)
			if err != nil {
				log.Printf("Error while unmarshaling: %v", err)
				continue
			}

			ch <- host

		}
		close(ch)
	}()

	return ch

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

			ch <- e

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

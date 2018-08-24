package main

import (
	"log"
	"os"
	"time"

	"github.com/rverton/webanalyze"
)

func Webanalyze() {

	filename := "/home/robin/work/rangemap/data/80-http-get-full_ipv4-20160905T233502.json"
	//filename := "/home/robin/work/rangemap/data/http-small.json"
	offlineAnalyze := true
	appsfile := "./apps.json"
	workers := 4
	verbose := true
	webanalyze.Timeout = 8 * time.Second

	log.Printf("Parsing IPs from '%v'", filename)
	log.Printf("Loading apps from '%v'", appsfile)

	checkAppsfile(appsfile)

	hosts := parseHttpHosts(filename)

	wa, err := webanalyze.NewWebAnalyzer(workers, appsfile)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for host := range hosts {
			var job *webanalyze.Job

			if offlineAnalyze {
				job = webanalyze.NewOfflineJob(
					host.Ip,
					host.Data.Http.Response.Body,
					host.Data.Http.Response.Headers)
			} else {
				job = webanalyze.NewOnlineJob(host.Ip, "", map[string][]string{})
			}

			wa.Schedule(job)
		}
		wa.Close()
	}()

	count := 0
	var t time.Time = time.Now()

	for result := range wa.Results {

		if count%1000 == 0 && count > 0 {
			log.Printf("%v processed (%v)", count, time.Since(t))
			t = time.Now()
		}

		if verbose {
			log.Printf("[%v]: %v", result.Duration, result.Host)
			for _, match := range result.Matches {
				log.Printf(" - %v, %v", match.AppName, match.Version)
			}
		}

		count++
	}
}

func checkAppsfile(appsfile string) error {
	if _, err := os.Stat(appsfile); os.IsNotExist(err) {
		log.Printf("Apps file '%v' was not found, downloading fresh one", appsfile)
		return webanalyze.DownloadFile(webanalyze.WappalyzerURL, appsfile)
	}

	return nil
}

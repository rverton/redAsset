package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/rverton/webanalyze"
)

var jsonOutput *json.Encoder

var wg sync.WaitGroup
var chInsert chan interface{}

func main() {

	parseCommand := flag.NewFlagSet("parse", flag.ExitOnError)
	analyzeCommand := flag.NewFlagSet("analyze", flag.ExitOnError)

	parseType := parseCommand.String("type", "rapid7-http", "File format. (rapid7-http|rapid7-fdns)")
	parseFile := parseCommand.String("file", "", "Filename to parse from. Gzip files allowed.")
	parseDomainFilter := parseCommand.String("domains", "", "File containing 2nd level domains to filter for.")
	parseDomainBlacklist := parseCommand.String("bdomains", "", "File containing 2nd level BLACKLIST domains.")
	parseOutput := parseCommand.String("output", "json", "Output format (json|postgres)")

	analyzeOutput := analyzeCommand.String("output", "postgres", "Output format (json|postgres)")
	analyzeInput := analyzeCommand.String("input", "postgres", "Input format (json|postgres)")
	analyzeWorker := analyzeCommand.Int("worker", 10, "Number of workers (default: 4)")
	analyzeAppsfile := analyzeCommand.String("appsfile", "./apps.json", "Apps.json definition file. Will be downloaded if not existent.")

	if len(os.Args) < 2 {
		fmt.Printf("Usage: %v <parse|analyze> [options]\n", os.Args[0])
		os.Exit(1)
	}

	switch os.Args[1] {

	case "parse":
		parseCommand.Parse(os.Args[2:])

	case "analyze":
		analyzeCommand.Parse(os.Args[2:])
	}

	jsonOutput = json.NewEncoder(os.Stdout)

	if parseCommand.Parsed() {

		var allowedDomains []string
		var blacklistDomains []string

		if *parseType == "" || *parseFile == "" {
			parseCommand.PrintDefaults()
			os.Exit(1)
		}

		if *parseDomainFilter != "" {
			var err error
			allowedDomains, err = parseDomainFile(*parseDomainFilter)
			if err != nil {
				log.Fatalf("Error reading domain file: %v", err)
			}

			log.Printf("Limiting to %v parsed domains.", len(allowedDomains))
		}

		if *parseDomainBlacklist != "" {
			var err error
			blacklistDomains, err = parseDomainFile(*parseDomainBlacklist)
			if err != nil {
				log.Fatalf("Error reading blacklist domain file: %v", err)
			}

			log.Printf("Limiting to %v parsed blacklist domains.", len(blacklistDomains))
		}

		var db *sql.DB
		var tt time.Time = time.Now()
		var count int = 1
		var valid int = 0
		chInsert = make(chan interface{})

		if *parseOutput == "postgres" {
			db = dbConnect()
			defer db.Close()
			go dbInsertWorker(db, chInsert)
		}

		switch *parseType {
		case "rapid7-http":

			hosts := parseHttpHosts(*parseFile)

			for host := range hosts {
				handleOutput(*parseOutput, host)
			}
		case "rapid7-fdns":

			dnsEntries := parseDnsHosts(*parseFile)

			for entry := range dnsEntries {

				progressStdout(count, valid, &tt, 1000000)

				count++

				// check if allowed domain
				if len(allowedDomains) > 0 {
					if !isAllowed(allowedDomains, entry.Name) {
						continue
					}
				}

				// remove blacklisted domains
				if len(blacklistDomains) > 0 {
					if isAllowed(blacklistDomains, entry.Name) {
						continue
					}
				}

				handleOutput(*parseOutput, entry)
				valid++
			}

		}

		close(chInsert)
		wg.Wait()
		log.Printf("Finished parsing.")
	}

	if analyzeCommand.Parsed() {

		if *analyzeInput == "" || *analyzeOutput == "" {
			parseCommand.PrintDefaults()
			os.Exit(1)
		}

		if *analyzeInput == "json" || *analyzeOutput == "json" {
			log.Fatalf("Analyzing from/to JSON not yet implemented.")
		}

		err := checkAppsfile(*analyzeAppsfile)
		if err != nil {
			log.Fatalf("Error retrieving apps definition: %v", err)
		}

		wa, err := webanalyze.NewWebAnalyzer(*analyzeWorker, *analyzeAppsfile)
		if err != nil {
			log.Fatal(err)
		}

		if *analyzeInput == "postgres" && *analyzeOutput == "postgres" {
			db := dbConnect()
			defer db.Close()

			rows, err := db.Query("SELECT hostname FROM hosts")
			if err != nil {
				log.Fatalf("Error querying data: %v", err)
			}

			go func() {
				var job *webanalyze.Job
				for rows.Next() {
					var hostname string
					err = rows.Scan(&hostname)
					if err != nil {
						log.Fatalf("Error querying data: %v", err)
					}

					job = webanalyze.NewOnlineJob(
						hostname,
						"",
						map[string][]string{})

					wa.Schedule(job)
				}

				wa.Close()
			}()

			count := 0
			tt := time.Now()

			for result := range wa.Results {
				count++

				progressStdout(count, count, &tt, 100)

				if result.Error != nil || len(result.Matches) <= 0 {
					continue
				}

				hostname := result.Host

				if strings.HasPrefix(hostname, "https://") {
					hostname = hostname[8:]
				}

				if strings.HasPrefix(hostname, "http://") {
					hostname = hostname[7:]
				}

				matches := map[string]string{}

				for _, m := range result.Matches {
					matches[m.AppName] = m.Version
				}

				data, err := json.Marshal(matches)
				if err != nil {
					log.Printf("Error marshaling: %v", err)
				}

				_, err = db.Exec(`UPDATE hosts
									SET apps = $1
								   WHERE
								    hostname = $2`,
					string(data),
					hostname)

				if err != nil {
					log.Printf("Error inserting: %v", err)
				}
			}
		}
	}

}

func progressStdout(count int, valid int, tt *time.Time, step int) {

	if count%step == 0 && count > 0 {
		log.Printf("\r%v processed, %v valid (%v)", count, valid, time.Since(*tt))
		*tt = time.Now()
	}

}

func handleOutput(output string, entry interface{}) {

	if output == "json" {
		jsonOutput.Encode(entry)
	}

	if output == "postgres" {
		wg.Add(1)
		chInsert <- entry
	}

}

func isAllowed(allowed []string, domain string) bool {

	for _, s := range allowed {
		if strings.HasSuffix(domain, s) {
			return true
		}
	}

	return false

}

func dbConnect() *sql.DB {
	var err error
	var db *sql.DB
	dbUri := os.Getenv("DB")

	if dbUri == "" {
		log.Println("Please specify a postgres uri via environment DB")
		log.Println("export DB=postgres://user:pw@host/db")
		os.Exit(1)
	}

	if db, err = sql.Open("postgres", dbUri); err != nil {
		log.Fatalf("Could not connect to postgres (host: %v): %v", dbUri, err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Could not connect to postgres (host: %v): %v", dbUri, err)
	}

	log.Println("Connected to database.")
	return db
}

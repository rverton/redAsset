# redasset

This tool allows to parse and analyze huge datasets to enumerate and analyze a target network/provider/company.

There are two main actions:

a) Parse scans.io gzipped files and import domains/hosts based on include/exclude 2nd domain lists.
b) Use a list of hosts to perform a webanalyze on them.

## Usage

### Configuration

To make use of postgres, export a DB URI:

    $ export DB=postgres://postgres:postgres@localhost/postgres?sslmode=disable

The docker-compose.yml file can be used to set up a postgres server.

### Parsing

Currently there two main methods implemented to parse hosts/domains from the scans.io project:

1. Parse a FDNS gzip file.
2. Parse a HTTP80 scan.

The following parameters may be used:

```
% ./redAsset parse -h
Usage of parse:
  -bdomains string
    	File containing 2nd level BLACKLIST domains.
  -domains string
    	File containing 2nd level domains to filter for.
  -file string
    	Filename to parse from. Gzip files allowed.
  -output string
    	Output format (json|postgres) (default "json")
  -type string
    	File format. (rapid7-http|rapid7-fdns) (default "rapid7-http")
```

Example of a FDNS parse:

    $ ./redAsset parse -file ~/work/rangemap/data/20170417-fdns.json.gz -domains dtag_second_level_domains.txt -bdomains dtag_second_level_domains_exclude.txt -output json -type rapid7-fdns
    2017/06/16 09:37:12 Limiting to 158 parsed domains.
    2017/06/16 09:37:12 Limiting to 19 parsed blacklist domains.
    {"Timestamp":"1492391616","Name":"0060118671.telekom-profis.de","Type":"a","Value":"80.237.195.199"}
    {"Timestamp":"1492392013","Name":"0060177055.telekom-profis.de","Type":"a","Value":"80.237.195.199"}
    {"Timestamp":"1492391637","Name":"0060242909.telekom-profis.de","Type":"a","Value":"80.237.195.199"}

### Analyzing

The analyze action will make use of github.com/rverton/webanalyze to guess used webapps (and their versions). The results can be exported (json) or directly written back to the databcan be exported (json) or directly written back to the database.

Example of a webanalyse from db:

    $ ./redAsset analyze -input postgres -output postgres

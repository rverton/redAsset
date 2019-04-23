# redasset

This tool allows to enumerate domains of a defined 2nd-lvl domain list by using the following methods:

1. Parse gzipped Rapid7 scans and filter by white- and blacklist.
2. Query certificate transparency logs.
3. Search for domains by passing CIDR.

The CIDR search can be used to map a huge list of IPs/networks to domain names.

## Usage

Note: The rapid7 file can still be gzipped.

The following parameters may be used:

```
$ ./redAsset -h
Usage of ./redAsset:
  -bdomains string
    	File containing 2nd level domains to exclude.
  -catransoff
    	Deactivate querying certificate transparency logs (crt.sh).
  -domains string
    	File containing 2nd level domains to include.
  -file string
    	JSON file to parse from, gzip allowed.
  -workers int
    	Number of workers to start. (default 4)


```

When CIDR notations are present, they are parsed and IPs of DNS entries are checked if they belong to these networks. Example `-domains` file:

    robinverton.de
    10.0.0.1/24

This will search for all domains ending with `.robinverton.de` and for DNS entries which map to an IP matching `10.0.0.1/24`.

Log is printed to *stderr*, results to *stdout*, so you may pipe this to a results file.

Example of a FDNS parse:

    $ ./redAsset -file ~/data/20170417-fdns.json.gz -domains second_level_domains.txt -bdomains blacklist_domains.txt
    2018/08/24 18:55:10 Limiting to 158 parsed domains.
    2018/08/24 18:55:10 Limiting to 19 parsed blacklist domains.
    0060118671.telekom-profis.de
    0060177055.telekom-profis.de
    [...]

### Post processing

Don't forget to deduplicate:

    $ cat ~/data/domains.txt | sort -u


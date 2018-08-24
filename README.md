# redasset

This tool allows to enumerate domains of a defined 2nd-lvl domain list by using the following methods:

1. Parse gzipped Rapid7 scans and filter by white- and blacklist.
2. Query certificate transparency logs.

## Usage

Note: The rapid7 file can still be gzipped.

The following parameters may be used:

```
$ Usage of ./redAsset:
  -bdomains string
      File containing 2nd level domains to exclude.
  -domains string
      File containing 2nd level domains to include.
  -file string
      Filename to parse from. Gzip files allowed.

```

Log is printed to stderr, results to stdout, so you may pipe this to a results file.

Example of a FDNS parse:

    $ ./redAsset -file ~/data/20170417-fdns.json.gz -domains second_level_domains.txt -bdomains blacklist_domains.txt -type rapid7-fdns
    2018/08/24 18:55:10 Limiting to 158 parsed domains.
    2018/08/24 18:55:10 Limiting to 19 parsed blacklist domains.
    0060118671.telekom-profis.de
    0060177055.telekom-profis.de
    [...]

### Post processing

Don't forget to deduplicate:

    $ cat ~/data/domains.txt | sort -u


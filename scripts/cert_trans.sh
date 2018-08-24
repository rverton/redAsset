#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 2nd-lvl-domains.txt"
    exit
fi

for domain in $(cat $1); do
    echo "Searching for $domain"
    curl "https://crt.sh/?q=%.$domain&output=json" | jq '.name_value' | sed 's/\"//g' | sed 's/\*\.//g' | sort -u
done

#!/usr/bin/env bash

set -o pipefail

sleep_the_clock_around() {
  sleep $((RANDOM % 60 + 10))
}

print_and_run() {
  echo "\$ ${*}"
  eval "${*}"
}

examples=(
  "spiry domain --bare --unix example.it"
  "spiry domain --bare example.it"
  "spiry domain --json example.com"
  "spiry domain --json --rfc3339 example.net"
  "spiry domain example.net"
  'spiry domain "example.shh"'
  'spiry domain "example.horse"'
)

for example in "${examples[@]}"; do
  print_and_run "${example}"
  sleep_the_clock_around
done

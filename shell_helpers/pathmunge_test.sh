#!/bin/sh

set -eu

helper=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)/pathmunge.sh

run_case() {
    name=$1
    shift
    output=$("$@" -c '
. "$1"
PATH=/a:/a:/a:/keep
pathmunge remove /a
first_status=$?
first=$PATH
pathmunge remove /a
second_status=$?
second=$PATH
pathmunge remove /a
third_status=$?
third=$PATH
pathmunge remove /a
fourth_status=$?
fourth=$PATH
printf "%s|%s|%s|%s|%s|%s|%s|%s\n" "$first" "$first_status" "$second" "$second_status" "$third" "$third_status" "$fourth" "$fourth_status"
' shell "$helper")
    expected="/a:/a:/keep|4|/a:/keep|4|/keep|4|/keep|5"
    if [ "$output" != "$expected" ]; then
        echo "$name: got $output, want $expected" >&2
        exit 1
    fi
}

run_case sh /bin/sh

if command -v bash >/dev/null 2>&1; then
    run_case bash bash --noprofile --norc
fi

if command -v zsh >/dev/null 2>&1; then
    run_case zsh zsh -f
fi

#!/bin/sh

set -e
rm -rf man
mkdir man
./szero manpage >"man/szero.1"
gzip man/szero.1

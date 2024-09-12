#!/bin/sh
# scripts/completions.sh
set -e
rm -rf completions
mkdir completions
for sh in bash zsh fish; do
	./szero completion "$sh" >"completions/szero.$sh"
done

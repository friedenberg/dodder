#! /bin/bash -e

ag "github.com/friedenberg/zit/\w+" "$@" -o --nofile --nocolor --nogroup | sort -u

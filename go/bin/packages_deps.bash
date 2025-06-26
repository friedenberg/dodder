#! /bin/bash -e

ag "github.com/friedenberg/dodder/\w+" "$@" -o --nofile --nocolor --nogroup | sort -u

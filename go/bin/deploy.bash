#! /bin/bash -e

./bin/lint.bash

git pull --rebase

#TODO pause mr-build-and-watch and then resume after
cmd_make="make"

if command -v gmake >/dev/null 2>&1; then
  cmd_make=gmake
else
  make
fi

$cmd_make build/deploy

git add .

if [[ "$(git status --porcelain=v1 2>/dev/null | wc -l)" -gt 0 ]]; then
  git commit -m update
fi

git push
# go clean -cache -fuzzcache

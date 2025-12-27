#! /usr/bin/env -S bash -e

zit show -format json :e |
  jq -s '{values: [.[]."object-id"]}' |
  yq -p=json -o=yaml

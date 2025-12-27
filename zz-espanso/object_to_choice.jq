#! /usr/bin/env -S jq -f -s

[
  .[]
  | {
    id: ."object-id",
    label: (
      if .description != "" then
        (."object-id" + ": " + .description)
      else
        ."object-id"
      end
    )
  }
]

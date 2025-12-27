#! /usr/bin/env -S jq -f

{
  matches: [
    {
      trigger: $ARGS.named.trigger,
      replace: "{{output}}",
      vars: [
        {
          name: "output",
          type: "choice",
          params: {
            values: .
          },
        }
      ],
    }
  ]
}

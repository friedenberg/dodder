#! /bin/bash -e

eval `/usr/libexec/path_helper -s`

cmd_args=()

for 

tacky copy $cmd_args \
  -i public.html - \
  -i public.utf8-plain-text <(./.zit/bin/strip-metadatei "$file" | pandoc -dforum-build --wrap=none)
echo "Copied to pasteboard"

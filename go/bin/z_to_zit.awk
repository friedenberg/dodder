#! /usr/bin/env awk -f

BEGIN {
  url = ""
  file = ""
  in_metadata = 1
}

FNR != 1 && $1 == "..." {
  print "---\n"
  in_metadata = 0
  next
}

/^- u-(https|file|chrome-extension):/ && in_metadata == 1 {
  url = $2
  gsub(/^u-/, "", url)
  next
}

/^- f-/ && in_metadata == 1 {
  file = $2
  gsub(/^f-/, "", file)
  print "! /Users/sasha/Zettelkasten/" file
  next
}

FNR == 2 {
  if (match($0, /^- (\S+\s|.*[^-a-z0-9_]+)/) == 0) {
    print "# "
  } else {
    $1 = "#"
  }
}

/^emp?t?y?$/ || /^..$/ {
  next
}

{
  print $0
}

END {
  if (url != "") {
    gsub(/^u-/, "", url)
    print url
    url = ""
  }
}

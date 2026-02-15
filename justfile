
dir_build := absolute_path("go/build")

default: build

#   ____        _ _     _
#  | __ ) _   _(_) | __| |
#  |  _ \| | | | | |/ _` |
#  | |_) | |_| | | | (_| |
#  |____/ \__,_|_|_|\__,_|
#

build:
  just go/build-go

#   _____         _
#  |_   _|__  ___| |_
#    | |/ _ \/ __| __|
#    | |  __/\__ \ |_
#    |_|\___||___/\__|
#

test-go *flags:
  just go/test-go-unit {{flags}}

test-bats: build test-bats-run

test-bats-run $PATH=(dir_build / "debug" + ":" + env("PATH")):
  just zz-tests_bats/test-generate_fixtures
  just zz-tests_bats/test

test-bats-targets *targets:
  #!/usr/bin/env bash
  export PATH="{{dir_build}}/debug:$PATH"
  just zz-tests_bats/test-targets {{targets}}

test-bats-tags *tags:
  #!/usr/bin/env bash
  export PATH="{{dir_build}}/debug:$PATH"
  just zz-tests_bats/test-tags {{tags}}

test: test-go test-bats

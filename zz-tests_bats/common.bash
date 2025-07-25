#! /bin/bash -e

if [[ -z $BATS_TEST_TMPDIR ]]; then
  echo 'common.bash loaded before $BATS_TEST_TMPDIR set. aborting.' >&2

  cat >&2 <<-'EOM'
    only load this file from `.bats` files like so:

    setup() {
      load "$(dirname "$BATS_TEST_FILE")/common.bash"

      # for shellcheck SC2154
      export output
    }

    as there is a hard assumption on $BATS_TEST_TMPDIR being set
EOM

  exit 1
fi

pushd "$BATS_TEST_TMPDIR" >/dev/null || exit 1

load "$BATS_CWD/test_helper/bats-support/load"
load "$BATS_CWD/test_helper/bats-assert/load"
load "$BATS_CWD/test_helper/bats-assert-additions/load"

# TODO remove this in favor of `-override-xdg-with-cwd`
set_xdg() {
  if [[ -z $1 ]]; then
    echo "trying to set empty XDG override. aborting." >&2
    exit 1
  fi

  loc="$(realpath "$1" 2>/dev/null)"

  if [[ -z $loc ]]; then
    echo "realpath for xdg is empty. aborting." >&2
    exit 1
  fi

  export XDG_DATA_HOME="$loc/.xdg/data"
  export XDG_CONFIG_HOME="$loc/.xdg/config"
  export XDG_STATE_HOME="$loc/.xdg/state"
  export XDG_CACHE_HOME="$loc/.xdg/cache"
  export XDG_RUNTIME_HOME="$loc/.xdg/runtime"
}

set_xdg "$BATS_TEST_TMPDIR"

# get the containing directory of this file
# use $BATS_TEST_FILENAME instead of ${BASH_SOURCE[0]} or $0,
# as those will point to the bats executable's location or the preprocessed file respectively
DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")" >/dev/null 2>&1 && pwd)"

cat_yin() (
  echo "one"
  echo "two"
  echo "three"
  echo "four"
  echo "five"
  echo "six"
)

cat_yang() (
  echo "uno"
  echo "dos"
  echo "tres"
  echo "quatro"
  echo "cinco"
  echo "seis"
)

cmd_dodder_def=(
  -debug no-tempdir-cleanup
  -abbreviate-zettel-ids=false
  -abbreviate-shas=false
  -predictable-zettel-ids
  -print-typen=false
  -print-time=false
  -print-etiketten=true
  -print-empty-shas=true
  -print-flush=false
  -print-unchanged=false
  -print-inventory_list=false
  -boxed-description=true
  -print-colors=false
)

export cmd_dodder_def

if [[ -z $DODDER_BIN ]]; then
  export DODDER_BIN
  echo "No \$DODDER_BIN set. This is usually set by .envrc or .env" >&2
  exit 1
fi

if [[ -z $DODDER_VERSION ]]; then
  export DODDER_VERSION
  DODDER_VERSION="v$("$DODDER_BIN" info store-version)"
fi

function copy_from_version {
  DIR="$1"
  rm -rf "$BATS_TEST_TMPDIR/.xdg"
  cp -r "$DIR/migration/$DODDER_VERSION/.xdg" "$BATS_TEST_TMPDIR/.xdg"
}

# TODO remove
function rm_from_version {
  chflags_and_rm
}

function chflags_and_rm {
  "$BATS_CWD/../bin/chflags.bash" -R nouchg "$BATS_TEST_TMPDIR"
}

function setup_repo {
  copy_from_version "$DIR" "$DODDER_VERSION"
}

function teardown_repo {
  chflags_and_rm
}

function run_dodder {
  cmd="$1"
  shift
  #shellcheck disable=SC2068
  run timeout --preserve-status "2s" "$DODDER_BIN" "$cmd" ${cmd_dodder_def[@]} "$@"
}

# TODO make this actually unify stderr
function run_dodder_stderr_unified {
  cmd="$1"
  shift
  #shellcheck disable=SC2068
  run "$DODDER_BIN" "$cmd" ${cmd_dodder_def[@]} "$@"
}

function run_dodder_init {
  if [[ $# -eq 0 ]]; then
    args=("test")
  else
    args=("$@")
  fi

  run_dodder init \
    -yin <(cat_yin) \
    -yang <(cat_yang) \
    -lock-internal-files=false \
    "${args[@]}"

  assert_success
  assert_output - <<-EOM
[!md @$(get_type_blob_sha) !toml-type-v1]
[konfig @$(get_konfig_sha) !toml-config-v1]
EOM

  run_dodder_init_workspace
}

function run_dodder_init_workspace {
  run_dodder init-workspace
}

function get_konfig_sha() {
  storeVersionCurrent="$(timeout --preserve-status "2s" "$DODDER_BIN" info "${cmd_dodder_def[@]}" store-version)"

  if [[ $storeVersionCurrent -le 10 ]]; then
    echo -n "9ad1b8f2538db1acb65265828f4f3d02064d6bef52721ce4cd6d528bc832b822"
  else
    echo -n "d23cb9e6237446e0ff798250c9e82862f29afd997581c9aefdf4916cebd00b90"
  fi
}

function get_type_blob_sha() {
  echo -n "b7ad8c6ccb49430260ce8df864bbf7d6f91c6860d4d602454936348655a42a16"
}

run_find() {
  run find . -maxdepth 2 ! -ipath './.xdg*' ! -iname '.dodder-workspace'
}

function run_dodder_init_disable_age {
  if [[ $# -eq 0 ]]; then
    args=("test-repo-id")
  else
    args=("$@")
  fi

  run_dodder init \
    -yin <(cat_yin) \
    -yang <(cat_yang) \
    -age-identity none \
    -lock-internal-files=false \
    "${args[@]}"

  assert_success
  assert_output - <<-EOM
[!md @$(get_type_blob_sha) !toml-type-v1]
[konfig @$(get_konfig_sha) !toml-config-v1]
EOM

  run_dodder blob_store-cat "$(get_konfig_sha)"
  assert_success
  assert_output

  run_dodder init-workspace
}

function start_server {
  dir="$1"

  coproc server {
    if [[ -n $dir ]]; then
      cd "$dir"
    fi

    # shellcheck disable=SC2068
    dodder serve ${cmd_dodder_def[@]} tcp :0
  }

  # shellcheck disable=SC2154
  # trap 'kill $server_PID' EXIT

  read -r output <&"${server[0]}"

  if [[ $output =~ (starting HTTP server on port: \"([0-9]+)\") ]]; then
    export port="${BASH_REMATCH[2]}"
  else
    fail <<-EOM
			unable to get port info from dodder server.
			server output: $output
		EOM
  fi
}

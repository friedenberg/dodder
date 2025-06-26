#! /bin/bash -e


dir_git_root="$(git rev-parse --show-toplevel)"
dir_base="$(realpath "$(dirname "$0")")"

v="$1"

if [[ -z "$1" ]]; then
  echo "no store version passed in" >&2
  exit 1
fi

d="${2:-$dir_base/v$v}"

if [[ -d $d ]]; then
  "$dir_git_root/bin/chflags.bash" -R nouchg "$d"
  rm -rf "$d"
fi

cmd_bats=(
  bats
  --tap
  --no-tempdir-cleanup
  migration/generate_fixture.bats
)

export BATS_TEST_TIMEOUT=3
if ! bats_run="$("${cmd_bats[@]}" 2>&1)"; then
  echo "$bats_run" >&2
  exit 1
else
  bats_dir="$(echo "$bats_run" | grep "BATS_RUN_TMPDIR" | cut -d' ' -f2)"
fi

mkdir -p "$d"
cp -r "$bats_dir/test/1/.xdg" "$d/.xdg"

#! /usr/bin/env -S bash -e
set -euo pipefail

if [[ "$(git branch --show-current)" != master ]]; then
  echo "not on master, refusing to deploy" >&2
  exit 1
fi

append_trap() {
  local sig="${2:-EXIT}"
  local cmd="$1"
  local existing=$(trap -p "$sig" | sed "s/.*'\(.*\)'.*/\1/")
  trap "${existing:+$existing; }$cmd" "$sig"
}

# commit worktree before codegen and tests
# TODO delete branch and reset commit if failure
function try() {
  {
    just ../zz-tests_bats/clean
    git add \
      . \
      ../{zz-pandoc,zz-vim,zz-tests_bats} \
      gomod2nix.toml

    git commit --allow-empty -m "update (pre codegen and test)"
  }

  target="deploy-github"
  timestamp="$(date +"%Y-%m-%d-%H-%M")"
  id="$target-$timestamp"
  worktree_dir="$(mktemp -d -t "$id-XXXXXX")"
  append_trap "rm -rf '$worktree_dir'"

  # make worktree
  {
    git branch "$id"
    append_trap "git branch -D '$id'"

    git worktree add "$worktree_dir" "$id"
    append_trap "git worktree remove --force '$worktree_dir'"

    # git diff HEAD | git -C "$worktree_dir" apply --3way

    prefix="$(git rev-parse --show-prefix)"

    pushd "$worktree_dir/$prefix" || exit 1
  }

  # modify worktree
  {
    just build-go codemod-go-fmt build-nix-gomod
  }

  # test worktree
  {
    to_test=(
      test
      # test-bats-next
    )

    just "${to_test[@]}"
  }
}

# commit worktree after codegen and tests
function commit_and_add() {
  {
    git add \
      . \
      ../{zz-pandoc,zz-vim,zz-tests_bats} \
      gomod2nix.toml

    git commit -m update
  }

  # merge worktree and push
  {
    popd
    git merge --autostash "$id"
    git push origin master
  }
}

try || git reset --soft HEAD^
commit_and_add

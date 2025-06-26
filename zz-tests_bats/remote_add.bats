#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	version="v$(zit info store-version)"
	copy_from_version "$DIR" "$version"
	run_zit_init_workspace
	export BATS_TEST_BODY=true

	# for shellcheck SC2154
	export output
}

teardown() {
	chflags_and_rm
}

# bats file_tags=user_story:remote

function print_their_xdg() {
	set_xdg "$(realpath "$1")"
	pushd "$1" >/dev/null || exit 1
	zit info-repo xdg
}

function remote_add_dotenv_xdg { # @test
	set_xdg them
	run_zit_init

	set_xdg "$BATS_TEST_TMPDIR"
	run_zit remote-add -remote-type native-dotenv-xdg <(print_their_xdg them) test-repo-id-them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/test-repo-id-them @[0-9a-f]+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_zit show /test-repo-id-them:k
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/test-repo-id-them @[0-9a-f]+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_zit show -format text /test-repo-id-them:k
	assert_success
	assert_output --regexp - <<-'EOM'
		---
		! toml-repo-dotenv_xdg-v0
		---

		public-key = 'zit-repo-public_key-v1.*'
		data = '/tmp/bats-run-\w+/test/.+/them/\.xdg/data/zit'
		config = '/tmp/bats-run-\w+/test/.+/\.xdg/config/zit'
		state = '/tmp/bats-run-\w+/test/.+/them/\.xdg/state/zit'
		cache = '/tmp/bats-run-\w+/test/.+/\.xdg/cache/zit'
		runtime = '/tmp/bats-run-\w+/test/.+/them/\.xdg/runtime/zit'
	EOM
}

function remote_add_local_path { # @test
	{
		set_xdg them
		mkdir -p them
		pushd them || exit 1
		run_zit_init -override-xdg-with-cwd test-repo-remote
		popd || exit 1
	}

	set_xdg "$BATS_TEST_TMPDIR"
	run_zit remote-add -remote-type stdio-local them test-repo-id-them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/test-repo-id-them @[0-9a-f]+ !toml-repo-local_path-v0]
	EOM

	run_zit show /test-repo-id-them:k
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/test-repo-id-them @[0-9a-f]+ !toml-repo-local_path-v0]
	EOM

	run_zit show -format text /test-repo-id-them:k
	assert_success
	assert_output --regexp - <<-'EOM'
		---
		! toml-repo-local_path-v0
		---

		public-key = 'zit-repo-public_key-v1.*'
		path = '/tmp/bats-run-\w+/test/.+/them'
	EOM
}

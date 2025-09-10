#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output

	copy_from_version "$DIR"
	run_dodder_init_workspace
}

teardown() {
	chflags_and_rm
}

# bats file_tags=user_story:remote

function print_their_xdg() {
	set_xdg "$(realpath "$1")"
	pushd "$1" >/dev/null || exit 1
	"$DODDER_BIN" info-repo xdg
}

function remote_add_dotenv_xdg { # @test
	set_xdg them
	run_dodder_init

	set_xdg "$BATS_TEST_TMPDIR"
	run_dodder remote-add -remote-type native-dotenv-xdg <(print_their_xdg them) test-repo-id-them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/test-repo-id-them @blake2b256-.+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_dodder show /test-repo-id-them:k
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/test-repo-id-them @blake2b256-.+ !toml-repo-dotenv_xdg-v0]
	EOM

	run_dodder show -format text /test-repo-id-them:k
	assert_success
	assert_output --regexp - <<-'EOM'
		---
		! toml-repo-dotenv_xdg-v0
		---

		public-key = 'dodder-repo-public_key-v1.*'
		data = '/tmp/bats-run-\w+/test/.+/them/\.xdg/data/dodder'
		config = '/tmp/bats-run-\w+/test/.+/\.xdg/config/dodder'
		state = '/tmp/bats-run-\w+/test/.+/them/\.xdg/state/dodder'
		cache = '/tmp/bats-run-\w+/test/.+/\.xdg/cache/dodder'
		runtime = '/tmp/bats-run-\w+/test/.+/them/\.xdg/runtime/dodder'
	EOM
}

function remote_add_local_path { # @test
	{
		set_xdg them
		mkdir -p them
		pushd them || exit 1
		run_dodder_init -override-xdg-with-cwd test-repo-remote
		popd || exit 1
	}

	set_xdg "$BATS_TEST_TMPDIR"
	run_dodder remote-add -remote-type stdio-local them test-repo-id-them
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/test-repo-id-them @blake2b256-.+ !toml-repo-local_path-v0]
	EOM

	run_dodder show /test-repo-id-them:k
	assert_success
	assert_output_unsorted --regexp - <<-'EOM'
		\[/test-repo-id-them @blake2b256-.+ !toml-repo-local_path-v0]
	EOM

	run_dodder show -format text /test-repo-id-them:k
	assert_success
	assert_output --regexp - <<-'EOM'
		---
		! toml-repo-local_path-v0
		---

		public-key = 'dodder-repo-public_key-v1.*'
		path = '/tmp/bats-run-\w+/test/.+/them'
	EOM
}

#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output
	export BATS_TEST_BODY=true
}

teardown() {
	chflags_and_rm
}

# bats file_tags=user_story:info

# bats test_tags=user_story:store_version
function info_store_version { # @test
	run_dodder info
	assert_output
}

# bats test_tags=user_story:compression
function info_compression_type { # @test
	run_dodder info compression-type
	assert_output 'zstd'
}

# bats test_tags=user_story:xdg
function info_xdg { # @test
	run_dodder_init_disable_age
	run_dodder info xdg
	assert_output - <<-EOM
		XDG_DATA_HOME=$BATS_TEST_TMPDIR/.xdg/data/dodder
		XDG_CONFIG_HOME=$BATS_TEST_TMPDIR/.xdg/config/dodder
		XDG_STATE_HOME=$BATS_TEST_TMPDIR/.xdg/state/dodder
		XDG_CACHE_HOME=$BATS_TEST_TMPDIR/.xdg/cache/dodder
		XDG_RUNTIME_HOME=$BATS_TEST_TMPDIR/.xdg/runtime/dodder
	EOM
}

function info_non_xdg { # @test
	run_dodder_init -override-xdg-with-cwd test-repo-id
	run_dodder info xdg
	assert_output - <<-EOM
		XDG_DATA_HOME=$BATS_TEST_TMPDIR/.dodder/local/share
		XDG_CONFIG_HOME=$BATS_TEST_TMPDIR/.dodder/config
		XDG_STATE_HOME=$BATS_TEST_TMPDIR/.dodder/local/state
		XDG_CACHE_HOME=$BATS_TEST_TMPDIR/.dodder/cache
		XDG_RUNTIME_HOME=$BATS_TEST_TMPDIR/.dodder/local/runtime
	EOM
}

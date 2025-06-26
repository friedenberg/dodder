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
	run_zit info
	assert_output
}

# bats test_tags=user_story:compression
function info_compression_type { # @test
	run_zit info compression-type
	assert_output 'zstd'
}

# bats test_tags=user_story:xdg
function info_xdg { # @test
	run_zit_init_disable_age
	run_zit info xdg
	assert_output - <<-EOM
		XDG_DATA_HOME=$BATS_TEST_TMPDIR/.xdg/data/zit
		XDG_CONFIG_HOME=$BATS_TEST_TMPDIR/.xdg/config/zit
		XDG_STATE_HOME=$BATS_TEST_TMPDIR/.xdg/state/zit
		XDG_CACHE_HOME=$BATS_TEST_TMPDIR/.xdg/cache/zit
		XDG_RUNTIME_HOME=$BATS_TEST_TMPDIR/.xdg/runtime/zit
	EOM
}

function info_non_xdg { # @test
	run_zit_init -override-xdg-with-cwd test-repo-id
	run_zit info xdg
	assert_output - <<-EOM
		XDG_DATA_HOME=$BATS_TEST_TMPDIR/.zit/local/share
		XDG_CONFIG_HOME=$BATS_TEST_TMPDIR/.zit/config
		XDG_STATE_HOME=$BATS_TEST_TMPDIR/.zit/local/state
		XDG_CACHE_HOME=$BATS_TEST_TMPDIR/.zit/cache
		XDG_RUNTIME_HOME=$BATS_TEST_TMPDIR/.zit/local/runtime
	EOM
}

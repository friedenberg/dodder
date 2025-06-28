#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	version="v$(dodder info store-version)"
	copy_from_version "$DIR" "$version"

	# for shellcheck SC2154
	export output
	export BATS_TEST_BODY=true
}

teardown() {
	chflags_and_rm
}

# bats file_tags=user_story:repo_info

# bats test_tags=user_story:config-immutable
function info_config_immutable { # @test
	run_dodder info store-version
	assert_success
	assert_output --regexp '[0-9]+'

	# shellcheck disable=SC2034
	storeVersionCurrent="$output"

	run_dodder info-repo config-immutable
	assert_success
	assert_output --regexp - <<-EOM
		---
		! toml-config-immutable-v1
		---

		public-key = 'dodder-repo-public_key-v1.*'
		store-version = $storeVersionCurrent
		repo-type = 'working-copy'
		id = 'test-repo-id'
		inventory_list-type = '!inventory_list-v2'

		\[blob-store]
		compression-type = 'zstd'
		lock-internal-files = false
	EOM
}

# bats test_tags=user_story:store_version
function info_store_version { # @test
	run_dodder info-repo
	assert_output
}

# bats test_tags=user_story:age_encryption
function info_age_none { # @test
	run_dodder info-repo age-encryption
	assert_output ''
}

# bats test_tags=user_story:age_encryption
function info_age_some { # @test
	age-keygen --output age-key >/dev/null 2>&1
	key="$(tail -n1 age-key)"
	run_dodder_init -override-xdg-with-cwd -age-identity age-key test-repo-id
	run_dodder info-repo age-encryption
	assert_output "$key"
}

# bats test_tags=user_story:compression
function info_compression_type { # @test
	run_dodder info-repo compression-type
	assert_output 'zstd'
}

# bats test_tags=user_story:xdg
function info_xdg { # @test
	run_dodder info-repo xdg
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
	run_dodder info-repo xdg
	assert_output - <<-EOM
		XDG_DATA_HOME=$BATS_TEST_TMPDIR/.dodder/local/share
		XDG_CONFIG_HOME=$BATS_TEST_TMPDIR/.dodder/config
		XDG_STATE_HOME=$BATS_TEST_TMPDIR/.dodder/local/state
		XDG_CACHE_HOME=$BATS_TEST_TMPDIR/.dodder/cache
		XDG_RUNTIME_HOME=$BATS_TEST_TMPDIR/.dodder/local/runtime
	EOM
}

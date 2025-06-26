#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	version="v$(zit info store-version)"
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
	run_zit info store-version
	assert_success
	assert_output --regexp '[0-9]+'

	# shellcheck disable=SC2034
	storeVersionCurrent="$output"

	run_zit info-repo config-immutable
	assert_success
	assert_output --regexp - <<-EOM
		---
		! toml-config-immutable-v1
		---

		public-key = 'zit-repo-public_key-v1.*'
		store-version = $storeVersionCurrent
		repo-type = 'working-copy'
		id = 'test-repo-id'

		\[blob-store]
		compression-type = 'zstd'
		lock-internal-files = false
	EOM
}

# bats test_tags=user_story:store_version
function info_store_version { # @test
	run_zit info-repo
	assert_output
}

# bats test_tags=user_story:age_encryption
function info_age_none { # @test
	run_zit info-repo age-encryption
	assert_output ''
}

# bats test_tags=user_story:age_encryption
function info_age_some { # @test
	age-keygen --output age-key >/dev/null 2>&1
	key="$(tail -n1 age-key)"
	run_zit_init -override-xdg-with-cwd -age-identity age-key test-repo-id
	run_zit info-repo age-encryption
	assert_output "$key"
}

# bats test_tags=user_story:compression
function info_compression_type { # @test
	run_zit info-repo compression-type
	assert_output 'zstd'
}

# bats test_tags=user_story:xdg
function info_xdg { # @test
	run_zit info-repo xdg
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
	run_zit info-repo xdg
	assert_output - <<-EOM
		XDG_DATA_HOME=$BATS_TEST_TMPDIR/.zit/local/share
		XDG_CONFIG_HOME=$BATS_TEST_TMPDIR/.zit/config
		XDG_STATE_HOME=$BATS_TEST_TMPDIR/.zit/local/state
		XDG_CACHE_HOME=$BATS_TEST_TMPDIR/.zit/cache
		XDG_RUNTIME_HOME=$BATS_TEST_TMPDIR/.zit/local/runtime
	EOM
}

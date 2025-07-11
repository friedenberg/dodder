#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output
}

teardown() {
	chflags_and_rm
}

function init_archive { # @test
	run_dodder info store-version
	assert_success
	assert_output --regexp '[0-9]+'

	# shellcheck disable=SC2034
	storeVersionCurrent="$output"

	run_dodder init-archive \
		-age-identity none \
		-lock-internal-files=false
	assert_success
	assert_output - <<-EOM
	EOM

	function output_immutable_config() {
		if [[ "$storeVersionCurrent" -le 10 ]]; then
			cat - <<-EOM
				---
				! toml-config-immutable-v1
				---

				public-key = 'dodder-repo-public_key-v1.*'
				store-version = $storeVersionCurrent
				repo-type = 'archive'
				id = ''
				inventory_list-type = '!inventory_list-v2'

				\[blob-store]
				compression-type = 'zstd'
				lock-internal-files = false
			EOM
		else
			cat - <<-EOM
				---
				! toml-config-immutable-v2
				---

				public-key = 'dodder-repo-public_key-v1.*'
				store-version = $storeVersionCurrent
				repo-type = 'archive'
				id = ''
				inventory_list-type = '!inventory_list-v2'
			EOM
		fi
	}

	run_dodder info-repo config-immutable
	assert_success
	output_immutable_config | assert_output --regexp -

	run_dodder cat-blob "$(get_konfig_sha)"
	assert_success
	assert_output

	run_dodder last
	assert_success
	assert_output ''

	run_dodder show
	assert_success
	assert_output ''
}

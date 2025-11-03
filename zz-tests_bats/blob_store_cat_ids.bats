#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output
}

function cat_ids { # @test
	run_dodder_init_disable_age
	assert_success

	run_dodder blob_store-cat-ids .default
	assert_success
	assert_output --partial "$(get_konfig_sha)"
}

#! /usr/bin/env bats

setup() {
	load "$(dirname "$BATS_TEST_FILE")/common.bash"

	# for shellcheck SC2154
	export output

	setup_repo
}

teardown() {
	teardown_repo
}

# bats file_tags=user_story:blob_store

function blob_store_sync_twice { # @test
	# TODO once migrated to madder blob stores for bats tests, enable this test again
	skip
	run_dodder blob_store-init test
	assert_success
	assert_output --regexp - <<-EOM
		Wrote config to .*/1-test.dodder-blob_store-config
	EOM

	run_dodder blob_store-sync
	assert_success
	assert_output --regexp - <<-EOM
		Successes: 14, Failures: 0, Ignored: 0, Total: 14
	EOM

	run_dodder blob_store-sync
	assert_success
	assert_output --regexp - <<-EOM
		Successes: 0, Failures: 0, Ignored: 14, Total: 14
	EOM
}
